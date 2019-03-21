// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package serv

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

//
func TestServer_Launch_1(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer func() { log.SetOutput(os.Stderr) }()

	server := NewServer(&http.Server{Addr: ":65535"})

	go func() {
		time.Sleep(time.Second)
		_ = server.Shutdown()
	}()
	if err := server.Launch(); err != nil {
		t.Error("unexpected error")
	}
}
