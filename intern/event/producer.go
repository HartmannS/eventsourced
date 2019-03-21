// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package event

import (
	"strings"
)

type (
	// Producer ...
	Producer interface {
		ServerSentEvent([]byte) ServerSentEvent
	}
	producer struct{}
)

var cleaner = strings.NewReplacer("\r\n", "\n", "\r", "")

// NewProducer ...
func NewProducer() Producer {
	return &producer{}
}

func (f *producer) ServerSentEvent(data []byte) ServerSentEvent {
	return newServerSentEvent(
		"",
		"",
		strings.Trim(cleaner.Replace(string(data)), " \n")+"\n",
		0,
	)
}
