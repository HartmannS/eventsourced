// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package app

import (
	"log"
	"net/http"
	"net/url"

	"eventsourced/intern/broker"
	"eventsourced/intern/conf"
	"eventsourced/intern/event"
	"eventsourced/intern/metric"
	"eventsourced/intern/serv"

	"github.com/streadway/amqp"
)

type (
	// Factory resolves dependencies.
	Factory interface {
		Server() serv.Server
	}
	factory struct {
		state  State
		metric metric.Metric
		brConn <-chan broker.Connection
	}
)

// NewFactory creates a new Factory form given State.
func NewFactory(state State, metric metric.Metric) Factory {
	return &factory{
		state:  state,
		metric: metric,
		brConn: yieldConn(state.Config(), metric),
	}
}

// Server ...
func (f *factory) Server() serv.Server {
	return serv.NewServer(
		&http.Server{
			Addr:    f.state.Config().Server.Address,
			Handler: f.serveMuxer(),
		},
	)
}

func (f *factory) serveMuxer() *http.ServeMux {
	muxer := http.NewServeMux()

	muxer.HandleFunc("/", f.endpoint)
	muxer.Handle("/debug/vars", http.DefaultServeMux)

	return muxer
}

func (f *factory) endpoint(w http.ResponseWriter, r *http.Request) {
	config := f.state.Config()
	pattern := serv.NewPattern(config.Queue.Pattern)
	expires := config.Queue.Expires

	producer := event.NewProducer()
	consumer := serv.NewConsumer(<-f.brConn, expires)
	defer func() { _ = consumer.Close() }()

	serv.NewServerSentHandler(
		consumer,
		pattern,
		producer,
		&serv.ResponseHeader{
			CORS: config.Header.CORS,
			SSE:  config.Header.SSE,
		},
		f.metric,
	).Handle(w, r)
}

func yieldConn(config conf.Config, metric metric.Metric) <-chan broker.Connection {
	var urls []*url.URL

	for _, v := range config.Broker.Node {
		if u, err := url.Parse(v); err != nil {
			log.Printf("%s", err)
		} else {
			urls = append(urls, u)
		}
	}

	return broker.NewConnector(amqp.Dial, urls, metric).Connection()
}
