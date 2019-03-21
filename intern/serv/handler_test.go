// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package serv

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"eventsourced/intern/event"
	"eventsourced/intern/metric"
	"eventsourced/intern/mock/serv"

	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
)

func TestServerSentEventHandler(t *testing.T) {
	NewServerSentHandler(nil, nil, nil, nil, nil)
}

// Must send HTTP 405 when method other than GET
func TestRequestHandler_Handle_1(t *testing.T) {
	unHandled := []string{
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodTrace,
	}

	for _, method := range unHandled {
		r := &handler{
			header: &ResponseHeader{},
		}
		recorder := httptest.NewRecorder()
		r.Handle(recorder, &http.Request{Method: method})

		if recorder.Code != 405 {
			t.Errorf("expected 405, got %d", recorder.Code)
		}
	}
}

// Must send HTTP 204 when OPTIONS requested
func TestRequestHandler_Handle_2(t *testing.T) {
	r := &handler{
		header: &ResponseHeader{},
	}
	recorder := httptest.NewRecorder()
	r.Handle(recorder, &http.Request{Method: "OPTIONS"})

	if recorder.Code != 204 {
		t.Errorf("expected 204, got %d", recorder.Code)
	}
}

// Must send HTTP 503 when queue name not determinable
func TestRequestHandler_Handle_3(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer func() { log.SetOutput(os.Stderr) }()

	r := &handler{
		pattern: NewPattern("${cookie:not-here}"),
		header:  &ResponseHeader{},
	}
	recorder := httptest.NewRecorder()
	r.Handle(recorder, &http.Request{Method: "GET"})

	if recorder.Code != 503 {
		t.Errorf("expected 503, got %d", recorder.Code)
	}
}

// Must send HTTP 503 when queue consumer fails
func TestRequestHandler_Handle_4(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer func() { log.SetOutput(os.Stderr) }()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock_serv.NewMockConsumer(ctrl)
	c.EXPECT().Consume(gomock.Any()).Return(nil, errors.New(""))

	r := &handler{
		consumer: c,
		pattern:  NewPattern("-"),
		header:   &ResponseHeader{},
	}

	recorder := httptest.NewRecorder()
	r.Handle(recorder, &http.Request{Method: "GET"})

	if recorder.Code != 503 {
		t.Errorf("expected 503, got %d", recorder.Code)
	}
}

type badWriter struct{}

func (w *badWriter) Header() http.Header        { return map[string][]string{} }
func (w *badWriter) Write([]byte) (int, error)  { return 0, errors.New("") }
func (w *badWriter) WriteHeader(statusCode int) {}
func (w *badWriter) Flush()                     {}

// Must continue on delivery failure and answer complete request
func TestRequestHandler_Handle_5(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	{
		d := make(chan amqp.Delivery, 2)
		d <- amqp.Delivery{Body: []byte("foo")}
		d <- amqp.Delivery{Body: []byte("bar")}

		c := mock_serv.NewMockConsumer(ctrl)
		c.EXPECT().Consume(gomock.Any()).Return(d, nil)
		c.EXPECT().Notify(gomock.Any())
		c.EXPECT().Ignore(gomock.Any())
		c.EXPECT().Close().Times(0)

		request := &http.Request{Method: "GET"}
		ctx, cancel := context.WithTimeout(request.Context(), time.Second)
		defer cancel()

		r := &handler{
			consumer: c,
			pattern:  NewPattern("-"),
			producer: event.NewProducer(),
			header:   &ResponseHeader{},
			metric:   metric.NewMetric("test"),
		}
		r.Handle(&badWriter{}, request.WithContext(ctx))
	}
	{
		d := make(chan amqp.Delivery, 2)
		d <- amqp.Delivery{Body: []byte("foo")}
		d <- amqp.Delivery{Body: []byte("bar")}

		c := mock_serv.NewMockConsumer(ctrl)
		c.EXPECT().Consume(gomock.Any()).Return(d, nil)
		c.EXPECT().Notify(gomock.Any())
		c.EXPECT().Ignore(gomock.Any())
		c.EXPECT().Close().Times(0)

		request := &http.Request{Method: "GET"}
		ctx, cancel := context.WithTimeout(request.Context(), time.Second)
		defer cancel()

		r := &handler{
			consumer: c,
			pattern:  NewPattern("-"),
			producer: event.NewProducer(),
			header:   &ResponseHeader{},
			metric:   metric.NewMetric("test"),
		}

		recorder := httptest.NewRecorder()
		r.Handle(recorder, request.WithContext(ctx))

		expect := ": SSE stream\n\ndata: foo\n\ndata: bar\n\n"
		result := recorder.Body.String()

		if expect != result {
			t.Errorf("unexpected response %s", result)
		}
	}
}

type hiccupConsumer struct{ d chan amqp.Delivery }

func (c *hiccupConsumer) Consume(string) (<-chan amqp.Delivery, error) {
	return c.d, nil
}
func (c *hiccupConsumer) Notify(err chan error) chan error {
	close(err)
	return err
}
func (c *hiccupConsumer) Ignore(chan error) {
}
func (c *hiccupConsumer) Close() error {
	return nil
}

// Must stop handling client requests on broker failure
func TestRequestHandler_Handle_6(t *testing.T) {
	h := &handler{
		consumer: &hiccupConsumer{},
		pattern:  NewPattern("-"),
		producer: event.NewProducer(),
		header:   &ResponseHeader{},
		metric:   metric.NewMetric("test"),
	}

	recorder := httptest.NewRecorder()
	h.Handle(recorder, &http.Request{Method: "GET"})

	// What about an assertion?
}
