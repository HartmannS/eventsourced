// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package app

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"eventsourced/intern/broker"
	"eventsourced/intern/conf"
	"eventsourced/intern/metric"
)

func TestFactory_Server(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer func() { log.SetOutput(os.Stderr) }()

	config := conf.NewConfig()
	config.Broker.Node[0] = "amqp://\\" // be broken

	version := NewVersion("", "", "", "")
	state := NewState(version, config)

	(&factory{state:state}).Server()
}

// Must dispatch to handler
func TestFactory_Muxer(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer func() { log.SetOutput(os.Stderr) }()

	version := NewVersion("", "", "", "")
	config := conf.NewConfig()
	state := NewState(version, config)

	f := &factory{
		state:  state,
		metric: metric.NewMetric(""),
		brConn: func() <-chan broker.Connection {
			ch := make(chan broker.Connection)
			close(ch)
			return ch
		}(),
	}

	endpoint, _ := url.Parse("/")
	recorder := httptest.NewRecorder()
	request := &http.Request{Method: "GET", URL: endpoint}

	f.serveMuxer().ServeHTTP(recorder, request)

	expect := "request parameter(s) missing"
	result := recorder.Result().Header["X-Status-Reason"][0]

	if expect != result {
		t.Errorf("unexpected response %s", result)
	}
}
