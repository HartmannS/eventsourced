// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package event

import (
	"testing"
)

// Must create string representation of SSE event
func TestServerSentEvent_String(t *testing.T) {
	samples := []struct{ given, expect string }{
		{given: "", expect: "data: \n"},
		{given: "x", expect: "data: x\n"},
		{given: "event: y\ndata: x", expect: "data: event: y\ndata: data: x\n"},
	}

	p := NewProducer()

	for _, sample := range samples {
		t.Run("", func(t *testing.T) {
			result := p.ServerSentEvent([]byte(sample.given))

			if result.String() != sample.expect {
				t.Errorf("expected %s, got %s", sample.expect, result.String())
			}
		})
	}
}

// Must have empty event id upon creation
func TestServerSentEvent_Id(t *testing.T) {
	if NewProducer().ServerSentEvent(nil).ID() != "" {
		t.Errorf("expected empty id")
	}
}

// Must have empty event type upon creation
func TestServerSentEvent_Event(t *testing.T) {
	if NewProducer().ServerSentEvent(nil).Event() != "" {
		t.Errorf("expected empty event type")
	}
}

// Must have retry rate set to 0 upon creation
func TestServerSentEvent_Retry(t *testing.T) {
	if NewProducer().ServerSentEvent(nil).Retry() != 0 {
		t.Errorf("expected retry to be 0")
	}
}
