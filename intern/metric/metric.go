// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package metric

import (
	"expvar"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type (
	// Metric ...
	Metric interface {
		Publish() Metric
		Report() interface{}

		IncBrokerCount()
		DecBrokerCount()

		IncConsumerCount()
		DecConsumerCount()

		IncDeliveryCount()
	}
	metric struct {
		sync.Mutex
		namespace string
		startTime time.Time
		broker    int64
		consumer  int64

		deliveryCount  int64
		deliverySample int64
	}

	// Stats ...
	Stats struct {
		Broker     int64
		Consumer   int64
		Delivery   int64
		Goroutines int
		Uptime     time.Duration
	}
)

// NewMetric ...
func NewMetric(namespace string) Metric {
	return &metric{namespace: namespace, startTime: time.Now().UTC()}
}

func (m *metric) Publish() Metric {
	expvar.Publish(m.namespace, expvar.Func(m.Report))

	go func() {
		for {
			time.Sleep(time.Second)
			m.Lock()
			m.deliverySample = m.deliveryCount
			m.Unlock()
			m.rstDeliveryCount()
		}
	}()

	return m
}

func (m *metric) Report() interface{} {
	m.Lock()
	defer m.Unlock()
	return Stats{
		Broker:     m.broker,
		Consumer:   m.consumer,
		Delivery:   m.deliverySample,
		Goroutines: runtime.NumGoroutine(),
		Uptime:     time.Since(m.startTime),
	}
}

func (m *metric) IncBrokerCount() { atomic.AddInt64(&m.broker, 1) }
func (m *metric) DecBrokerCount() { atomic.AddInt64(&m.broker, -1) }

func (m *metric) IncConsumerCount() { atomic.AddInt64(&m.consumer, 1) }
func (m *metric) DecConsumerCount() { atomic.AddInt64(&m.consumer, -1) }

func (m *metric) IncDeliveryCount() { atomic.AddInt64(&m.deliveryCount, 1) }
func (m *metric) rstDeliveryCount() { atomic.StoreInt64(&m.deliveryCount, 0) }
