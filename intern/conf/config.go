// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package conf

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type (
	// Server ...
	Server struct {
		Address string `yaml:"address"`
	}
	// Broker ...
	Broker struct {
		Node []string `yaml:"node"`
	}
	// Queue ...
	Queue struct {
		Pattern string `yaml:"pattern"`
		Expires int    `yaml:"expires"`
	}
	// Header ...
	Header struct {
		CORS map[string]string `yaml:"cors"`
		SSE  map[string]string `yaml:"sse"`
	}

	// Config ...
	Config struct {
		Server Server `yaml:"server"`
		Broker Broker `yaml:"broker"`
		Queue  Queue  `yaml:"queue"`
		Header Header `yaml:"header"`
		source []string
		loaded bool
	}
)

func defaults() *Config {
	return &Config{
		Server: Server{
			Address: "0.0.0.0:2069",
		},
		Broker: Broker{
			Node: []string{
				"amqp://guest:guest@127.0.0.1:5672/", // channelMax 2047
				"amqp://guest:guest@127.0.0.1:5672/", // channelMax 2047
			},
		},
		Queue: Queue{
			Pattern: "${query:id}",
			Expires: 1800,
		},
		Header: Header{
			CORS: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Accept, Cache-Control",
			},
			SSE: map[string]string{
				"Content-Type":      "text/event-stream; charset=utf-8",
				"Cache-Control":     "no-cache",
				"Transfer-Encoding": "identity",
				"X-Accel-Buffering": "no",
			},
		},
	}
}

// NewConfig ...
func NewConfig(filename ...string) Config {
	var bytes []byte
	var err error

	config := defaults()

	for _, file := range filename {
		if bytes, err = ioutil.ReadFile(file); err == nil {
			if err = yaml.Unmarshal(bytes, config); err != nil {
				fmt.Print(err)
				continue
			}
			config.source = append(config.source, file)
			config.loaded = true
		}
	}

	return *config
}

// Source ...
func (c Config) Source() []string {
	return c.source
}

// Loaded ...
func (c Config) Loaded() bool {
	return c.loaded
}
