// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package serv

import (
	"fmt"
	"log"
	"net/http"

	"eventsourced/intern/event"
	"eventsourced/intern/metric"

	"github.com/streadway/amqp"
)

type (
	// ResponseHandler ...
	ResponseHandler interface {
		Handle(w http.ResponseWriter, r *http.Request)
	}
	handler struct {
		consumer Consumer
		producer event.Producer
		pattern  Pattern
		header   *ResponseHeader
		metric   metric.Metric
	}

	// ResponseHeader ...
	ResponseHeader struct {
		CORS map[string]string
		SSE  map[string]string
	}
)

// NewServerSentHandler ...
func NewServerSentHandler(
	consumer Consumer,
	pattern Pattern,
	producer event.Producer,
	header *ResponseHeader,
	metric metric.Metric,
) ResponseHandler {
	return &handler{
		consumer: consumer,
		pattern:  pattern,
		producer: producer,
		header:   header,
		metric:   metric,
	}
}

// Handle ...
func (h *handler) Handle(w http.ResponseWriter, r *http.Request) {
	var err error
	var queue string
	var messages <-chan amqp.Delivery

	if r.Method == "OPTIONS" {
		h.sendStatus(w, http.StatusNoContent, nil)
		return
	}
	if r.Method != "GET" {
		h.sendStatus(w, http.StatusMethodNotAllowed, nil)
		return
	}
	if queue, err = h.pattern.Apply(r); err != nil {
		h.sendStatus(w, http.StatusServiceUnavailable, err)
		return
	}
	if messages, err = h.consumer.Consume(queue); err != nil {
		h.sendStatus(w, http.StatusServiceUnavailable, err)
		return
	}

	h.setHeader(w, h.header.SSE)
	h.setHeader(w, h.header.CORS)
	h.sendBanner(w)

	brokerClose := h.consumer.Notify(make(chan error))
	defer h.consumer.Ignore(brokerClose)

	clientClose := r.Context().Done()

	deliver := func(msg []byte) error {
		ev := h.producer.ServerSentEvent(msg).String()
		if _, err := fmt.Fprintln(w, ev); err != nil {
			return err
		}
		w.(http.Flusher).Flush()
		return nil
	}

	for {
		select {
		case message := <-messages:
			if err := deliver(message.Body); err != nil {
				continue
			}
			_ = message.Ack(false)
			h.metric.IncDeliveryCount()

		case _ = <-brokerClose:
			return
		case _ = <-clientClose:
			return
		}
	}
}

func (h *handler) setHeader(w http.ResponseWriter, header map[string]string) {
	for k, v := range header {
		w.Header().Set(k, v)
	}
}

func (h *handler) sendStatus(w http.ResponseWriter, status int, err error) {
	if status >= 400 && err != nil {
		h.setHeader(w, map[string]string{
			"X-Status-Reason": err.Error(),
		})
		log.Printf("server: status %d, %s", status, err.Error())
	}

	h.setHeader(w, h.header.CORS)
	w.WriteHeader(status)
}

func (h *handler) sendBanner(w http.ResponseWriter) {
	_, _ = fmt.Fprintf(w, ": SSE stream\n\n")
	w.(http.Flusher).Flush()
}
