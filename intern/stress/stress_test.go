// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package stress

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"eventsourced/intern/app"
	"eventsourced/intern/conf"
	"eventsourced/intern/metric"

	"github.com/streadway/amqp"
)

var testStress = struct {
	queueName   string
	numQueues   int
	numMessages int
	ctxTimeout  time.Duration
}{
	queueName:   "eventsourced-stress-test-debris.delete.me",
	numQueues:   1000, // cave: ulimit -n
	numMessages: 42,
	ctxTimeout:  time.Second * 5,
}

func TestStress(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer func() { log.SetOutput(os.Stderr) }()

	state := testStressState()

	factory := app.NewFactory(state, metric.NewMetric("eventsourced"))
	service := factory.Server()

	go service.Launch()
	defer service.Shutdown()

	testStressPrepare(t)
	defer testStressCleanUp(t)

	time.Sleep(time.Second * 2)
	testStressExecute(t)
}

func testStressPrepare(t *testing.T) {
	var err error
	var conn *amqp.Connection

	url := testStressConfig().Broker.Node[0]
	if conn, err = amqp.Dial(url); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	ch, _ := conn.Channel()
	defer func() { _ = ch.Close() }()

	exp := testStressConfig().Queue.Expires
	args := amqp.Table{"x-expires": int32(exp * 1000)}

	for i := 0; i < testStress.numQueues; i++ {
		queue := testStress.queueName + "-" + strconv.Itoa(i)
		_, _ = ch.QueueDeclare(queue, true, false, false, false, args)

		for j := 0; j < testStress.numMessages; j++ {
			pub := amqp.Publishing{Body: []byte(strconv.Itoa(j + 1))}
			_ = ch.Publish("", queue, false, false, pub)
		}
	}
}

func testStressCleanUp(t *testing.T) {
	var err error
	var conn *amqp.Connection

	url := testStressConfig().Broker.Node[0]
	if conn, err = amqp.Dial(url); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	ch, _ := conn.Channel()
	defer func() { _ = ch.Close() }()

	for i := 0; i < testStress.numQueues; i++ {
		queue := testStress.queueName + "-" + strconv.Itoa(i)
		if _, err := ch.QueueDelete(queue, false, false, false); err != nil {
			t.Error(err)
		}
	}
}

func testStressExecute(t *testing.T) {
	// Sum of all message payload numbers: 1 + 2 + 3 + ... + n
	expect := (testStress.numMessages * (testStress.numMessages + 1)) / 2

	matcher := regexp.MustCompile(`\ndata:\s*(\d+)`)

	wg := &sync.WaitGroup{}
	for i := 0; i < testStress.numQueues; i++ {

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			var res *http.Response
			var err error
			var body []byte

			addr := testStressConfig().Server.Address
			queue := testStress.queueName
			index := strconv.Itoa(i)

			url := fmt.Sprintf("http://%s/?id=%s-%s", addr, queue, index)

			ctx, cancel := context.WithTimeout(
				context.Background(),
				testStress.ctxTimeout,
			)
			defer cancel()

			req, _ := http.NewRequest("GET", url, nil)
			req = req.WithContext(ctx)

			if res, err = (&http.Client{}).Do(req); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = res.Body.Close() }()

			body, _ = ioutil.ReadAll(io.Reader(res.Body))
			all := matcher.FindAllSubmatch(body, -1)

			result := 0
			for _, v := range all {
				if x, err := strconv.Atoi(string(v[1])); err == nil {
					result = result + x
				}
			}
			if result != expect {
				t.Errorf(
					"expected %d, got %d, failure: %s",
					expect,
					result,
					res.Header.Get("x-status-reason"),
				)
			}
		}(i)
	}
	wg.Wait()
}

func testStressState() app.State {
	return app.NewState(app.NewVersion("", "", "", ""), testStressConfig())
}

func testStressConfig() conf.Config {
	return conf.Config{
		Server: conf.Server{
			Address: "127.0.0.1:12069",
		},
		Broker: conf.Broker{
			Node: []string{
				"amqp://guest:guest@127.0.0.1:5672/", // channelMax 2047
				"amqp://guest:guest@127.0.0.1:5672/", // channelMax 2047
				"amqp://guest:guest@127.0.0.1:5672/", // channelMax 2047
			},
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
