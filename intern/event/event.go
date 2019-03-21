// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package event

import (
	"strings"
)

type (
	// ServerSentEvent ...
	ServerSentEvent interface {
		ID() string
		Event() string
		Data() string
		Retry() int
		String() string
	}
	sseEvent struct {
		id    string
		event string
		data  string
		retry int
	}
)

func newServerSentEvent(id, kind, data string, retry int) ServerSentEvent {
	return &sseEvent{id, kind, data, retry}
}

func (e *sseEvent) ID() string    { return e.id }
func (e *sseEvent) Event() string { return e.event }
func (e *sseEvent) Data() string  { return e.data }
func (e *sseEvent) Retry() int    { return e.retry }

func (e *sseEvent) String() string {
	s := strings.Trim(e.data, "\n")
	return "data: " + strings.Replace(s, "\n", "\ndata: ", -1) + "\n"
}
