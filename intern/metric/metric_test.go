// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package metric

import (
	"sync"
	"testing"
	"time"
)

func TestMetric_Publish(t *testing.T) {
	m := NewMetric("")

	m.IncDeliveryCount()
	m.IncDeliveryCount()
	m.IncDeliveryCount()
	m.Publish()

	expect := int64(0)
	result := m.Report().(Stats).Delivery

	if result != expect {
		t.Errorf("expected %d, got %d", expect, result)
	}

	time.Sleep(time.Second)
	time.Sleep(time.Millisecond * 500)

	expect = int64(3)
	result = m.Report().(Stats).Delivery

	if result != expect {
		t.Errorf("expected %d, got %d", expect, result)
	}
}

// Must count metric, must not cause a data race
func TestMetric_BrokerCount(t *testing.T) {
	m := NewMetric("")

	wg := &sync.WaitGroup{}
	for i := 0; i < 42; i++ {
		wg.Add(1)
		go func() { m.IncBrokerCount(); wg.Done() }()

		if i%10 == 0 {
			wg.Add(1)
			go func() { m.DecBrokerCount(); wg.Done() }()
		}
	}
	wg.Wait()

	expect := int64(37)
	result := m.Report().(Stats).Broker

	if result != expect {
		t.Errorf("expected %d, got %d", expect, result)
	}
}

// Must count metric, must not cause a data race
func TestMetric_ConsumerCount(t *testing.T) {
	m := NewMetric("")

	wg := &sync.WaitGroup{}
	for i := 0; i < 42; i++ {
		wg.Add(1)
		go func() { m.IncConsumerCount(); wg.Done() }()

		if i%10 == 0 {
			wg.Add(1)
			go func() { m.DecConsumerCount(); wg.Done() }()
		}
	}
	wg.Wait()

	expect := int64(37)
	result := m.Report().(Stats).Consumer

	if result != expect {
		t.Errorf("expected %d, got %d", expect, result)
	}
}
