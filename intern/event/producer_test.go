// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package event

import (
	"testing"
)

// Must create event with normalized data
func TestProducer_Event(t *testing.T) {
	samples := []struct{ given, expect string }{
		{given: "" /*      */, expect: "\n"},
		{given: "           ", expect: "\n"},
		{given: " f\noo \n\n", expect: "f\noo\n"},
		{given: " :data     ", expect: ":data\n"},
		{given: "X\r\nX\n\r ", expect: "X\nX\n"},
	}

	p := NewProducer()

	for _, sample := range samples {
		t.Run("", func(t *testing.T) {
			result := p.ServerSentEvent([]byte(sample.given))

			if result.Data() != sample.expect {
				t.Errorf("expected %s, got %s", sample.expect, result.Data())
			}
		})
	}
}
