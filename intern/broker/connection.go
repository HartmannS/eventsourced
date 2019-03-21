// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package broker

import (
	"errors"
	"sync"

	"github.com/streadway/amqp"
)

type (
	// Connection represents an observable connection to an AMQP broker.
	Connection interface {
		Channel() (*amqp.Channel, error)
		Notify(err chan error) chan error
		Ignore(err chan error)
		Close() error
	}
	connection struct {
		conn *amqp.Connection

		mu  sync.Mutex
		err map[chan error]bool
	}
)

func newConnection(conn *amqp.Connection) Connection {
	p := &connection{conn: conn, err: map[chan error]bool{}}
	go func() { p.dispatch(<-conn.NotifyClose(make(chan *amqp.Error))) }()
	return p
}

// Channel ...
func (p *connection) Channel() (*amqp.Channel, error) {
	return p.conn.Channel()
}

// Close ...
func (p *connection) Close() error {
	return p.conn.Close()
}

// Notify ...
func (p *connection) Notify(err chan error) chan error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.err[err] = true
	return err
}

// Ignore ...
func (p *connection) Ignore(err chan error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.err, err)
}

func (p *connection) dispatch(err *amqp.Error) {
	if err != nil {
		p.mu.Lock()
		defer p.mu.Unlock()

		for listener := range p.err {
			listener <- errors.New(err.Error())
		}
	}
}
