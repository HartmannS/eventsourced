// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package app

import (
	"eventsourced/intern/conf"
)

type (
	// State ...
	State interface {
		Version() Version
		Config() conf.Config
	}
	state struct {
		version Version
		config  conf.Config
	}
)

// NewState ...
func NewState(version Version, config conf.Config) State {
	return &state{
		version: version,
		config:  config,
	}
}

func (s *state) Version() Version {
	return s.version
}

func (s *state) Config() conf.Config {
	return s.config
}
