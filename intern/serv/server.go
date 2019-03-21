// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package serv

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type (
	// Server ...
	Server interface {
		Launch() error
		Shutdown() error
	}
	server struct {
		server *http.Server
	}
)

const (
	logServerListen = "server: listening %s"
	logServerClosed = "server: closed %s"
	logServerSignal = "server: caught signal: %s (%#v)"
	logServerError  = "server: %s"
)

// NewServer ...
func NewServer(srv *http.Server) Server {
	return &server{srv}
}

// Launch ...
func (s *server) Launch() error {
	served := make(chan error)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	log.Printf(logServerListen, s.server.Addr)
	go func() { served <- s.server.ListenAndServe() }()

	select {
	case err := <-served:
		if err == http.ErrServerClosed {
			log.Printf(logServerClosed, s.server.Addr)
			return nil
		}
		log.Printf(logServerError, err)
		return err

	case catch := <-sig:
		signal.Stop(sig)
		log.Printf(logServerSignal, catch, catch)
		return s.Shutdown()
	}
}

// Shutdown ...
func (s *server) Shutdown() error {
	emptyCtx := context.Background()
	duration := time.Second * 2

	ctx, cancel := context.WithTimeout(emptyCtx, duration)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf(logServerError, err)
		cancel()
	}
	<-ctx.Done()
	return nil
}
