// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package stress

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"eventsourced/intern/app"
	"eventsourced/intern/conf"
	"eventsourced/intern/metric"

	"github.com/streadway/amqp"
)

func TestSmoke(t *testing.T) {
	var err error
	var res *http.Response

	log.SetOutput(ioutil.Discard)
	defer func() { log.SetOutput(os.Stderr) }()

	expect := ": SSE stream\n"
	expect += "\n"
	expect += "data: { do not touch whitespaces }\n"
	expect += "\n"
	expect += "data: { in this fixture }\n"
	expect += "\n"
	expect += "data: { odd message\n"
	expect += "data: data: !\n"
	expect += "data:  :\n"
	expect += "data: ! }\n"
	expect += "\n"

	state := testSmokeState()
	factory := app.NewFactory(state, metric.NewMetric("eventsourced"))
	service := factory.Server()

	go service.Launch()
	defer service.Shutdown()

	broker := state.Config().Broker.Node[0]
	queue := "eventsourced-smoke-test-debris.delete.me"

	{
		channel, err := testSmokeChannel(broker)
		if err != nil {
			t.Skipf("[%s] %s", broker, err)
		}

		exp := state.Config().Queue.Expires
		args := amqp.Table{"x-expires": int32(exp * 1000)}

		_, err = channel.QueueDeclare(queue, true, false, false, false, args)
		if err != nil {
			t.Fatalf("%s", err)
		}
		defer func() { _, _ = channel.QueueDelete(queue, false, false, false) }()

		msg0 := amqp.Publishing{Body: []byte("{ do not touch whitespaces }")}
		msg1 := amqp.Publishing{Body: []byte("{ in this fixture }")}
		msg2 := amqp.Publishing{Body: []byte("{ odd message\ndata: !\n :\n! }")}

		_ = channel.Publish("", queue, false, false, msg0)
		_ = channel.Publish("", queue, false, false, msg1)
		_ = channel.Publish("", queue, false, false, msg2)
	}
	{
		client := &http.Client{Timeout: time.Second}
		addr := state.Config().Server.Address
		url := fmt.Sprintf("http://%s/?id=%s", addr, queue)

		if res, err = client.Get(url); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = res.Body.Close() }()

		result, _ := ioutil.ReadAll(res.Body)

		if !bytes.Equal(result, []byte(expect)) {
			t.Errorf("response body and fixture are not equal")
		}
	}

}

func testSmokeState() app.State {
	return app.NewState(app.NewVersion("", "", "", ""), testSmokeConfig())
}

func testSmokeConfig() conf.Config {
	return conf.Config{
		Server: conf.Server{
			Address: "127.0.0.1:12069",
		},
		Broker: conf.Broker{
			Node: []string{"amqp://guest:guest@127.0.0.1:5672/"},
		},
		Queue: conf.Queue{
			Pattern: "${query:id}",
			Expires: 1800,
		},
		Header: conf.Header{
			CORS: map[string]string{"X-CORS": "CORS"},
			SSE:  map[string]string{"X-SSE": "SSE"},
		},
	}
}

func testSmokeChannel(url string) (*amqp.Channel, error) {
	broker, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	channel, err := broker.Channel()
	if err != nil {
		return nil, err
	}
	return channel, nil
}
