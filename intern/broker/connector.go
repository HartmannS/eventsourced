// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package broker

import (
	"log"
	"net/url"
	"time"

	"eventsourced/intern/metric"

	"github.com/streadway/amqp"
)

type (
	// Connector provides AMQP connection on demand.
	Connector interface {
		Connection() <-chan Connection
	}
	connector struct {
		dialer  Dialer
		brokers []*url.URL
		metric  metric.Metric
	}

	// Dialer is the callback function that provides an AMQP connection.
	Dialer = func(url string) (*amqp.Connection, error)
)

// NewConnector creates a new broker connector.
func NewConnector(
	dialer Dialer,
	brokers []*url.URL,
	metric metric.Metric,
) Connector {
	return &connector{dialer, brokers, metric}
}

// Connection yields a broker connection on demand
func (p *connector) Connection() <-chan Connection {
	yield := make(chan Connection)

	provide := func(brokerURL *url.URL) {
		var brConn Connection
		var onErr chan error

		for {
			if brConn == nil {
				amqpConn, err := p.dialer(brokerURL.String())

				if err != nil {
					log.Printf("dialer: %s", err)
					time.Sleep(5 * time.Second)
					continue
				}

				brConn = newConnection(amqpConn)
				onErr = brConn.Notify(make(chan error))
				connected(brokerURL, p.metric)
			}

			select {
			case err := <-onErr:
				brConn = nil
				disconnected(brokerURL, err, p.metric)
				continue

			default:
				yield <- brConn
			}
		}
	}

	for _, brokerURL := range p.brokers {
		go provide(brokerURL)
	}

	// detect connection loss when idle
	go func() {
		for {
			<-yield
			time.Sleep(1 * time.Second)
		}
	}()

	return yield
}

func connected(brokerURL *url.URL, metric metric.Metric) {
	metric.IncBrokerCount()
	log.Printf("broker: connected %s", sanitize(brokerURL))
}

func disconnected(brokerURL *url.URL, err error, metric metric.Metric) {
	metric.DecBrokerCount()
	if err != nil {
		log.Printf("broker: closing %s", err)
	}
	log.Printf("broker: disconnected %s", sanitize(brokerURL))
}

func sanitize(url *url.URL) string {
	u, _ := url.Parse(url.String())
	u.Scheme, u.Opaque = "", u.Host
	return u.String()
}
