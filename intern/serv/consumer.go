// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package serv

//go:generate mockgen -destination=../mock/serv/consumer.go eventsourced/intern/serv Consumer

import (
	"errors"
	"sync"

	"eventsourced/intern/broker"

	"github.com/streadway/amqp"
)

type (
	// Consumer ...
	Consumer interface {
		Consume(queue string) (<-chan amqp.Delivery, error)
		Notify(chan error) chan error
		Ignore(chan error)
		Close() error
	}

	consumer struct {
		conn    broker.Connection
		expires int

		mu sync.Mutex
		ch *amqp.Channel
	}
)

// NewConsumer ...
func NewConsumer(conn broker.Connection, expires int) Consumer {
	return &consumer{conn: conn, expires: expires}
}

// Consume ...
func (c *consumer) Consume(name string) (<-chan amqp.Delivery, error) {
	var err error
	var ch *amqp.Channel
	var q amqp.Queue

	var args = map[string]interface{}{
		"x-expires": int32(c.expires * 1000),
	}

	if ch, err = c.conn.Channel(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.ch = ch
	c.mu.Unlock()

	if q, err = c.ch.QueueDeclare(
		name,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		args,
	); err != nil {
		return nil, err
	}

	if q.Consumers != 0 {
		return nil, errors.New("server: max consumers exceeded")
	}
	if err = c.ch.Qos(1, 0, false); err != nil {
		return nil, err
	}
	if err = c.ch.Confirm(false); err != nil {
		return nil, err
	}

	return c.ch.Consume(
		name,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		args,
	)
}

// Notify ...
func (c *consumer) Notify(err chan error) chan error {
	return c.conn.Notify(err)
}

// Ignore ...
func (c *consumer) Ignore(err chan error) {
	c.conn.Ignore(err)
}

// Close ...
func (c *consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
	}
	return nil
}
