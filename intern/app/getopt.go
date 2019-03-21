// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package app

import (
	"flag"
	"fmt"

	"gopkg.in/yaml.v2"
)

type (
	// GetOpt ...
	GetOpt interface {
		Complement(s State) bool
	}
	getOpt struct {
		version  *bool
		confdump *bool
	}
)

// NewGetOpt ...
func NewGetOpt() GetOpt {
	options := &getOpt{
		version:  flag.Bool("v", false, "display version"),
		confdump: flag.Bool("c", false, "dump configuration"),
	}
	flag.Parse()
	return options
}

func (o *getOpt) Complement(s State) bool {
	if *o.version {
		fmt.Println(s.Version().Long())
		return false
	}

	if *o.confdump {
		j, _ := yaml.Marshal(s.Config())
		fmt.Printf("%s", j)
		return false
	}

	// ...

	return true
}
