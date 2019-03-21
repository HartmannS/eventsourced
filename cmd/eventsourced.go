// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package main

import (
	_ "expvar"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"eventsourced/intern/app"
	"eventsourced/intern/conf"
	"eventsourced/intern/metric"
)

var (
	project = ""
	version = ""
	commit  = ""
	status  = ""
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix(fmt.Sprintf("[%d] ", os.Getpid()))

	var (
		version = app.NewVersion(project, version, commit, status)
		config  = loadConfig()
		state   = app.NewState(version, config)
		metrics = metric.NewMetric(project).Publish()
	)

	if !app.NewGetOpt().Complement(state) {
		return
	}

	if config.Loaded() {
		configs := strings.Join(config.Source()[:], ", ")
		log.Printf("config: loaded %s", configs)
	} else {
		log.Print("config: not loaded, using default preset")
	}

	_ = app.NewFactory(state, metrics).Server().Launch()
}

func loadConfig() conf.Config {
	var selfFile string
	var confPath string
	var err error

	if selfFile, err = os.Executable(); err != nil {
		panic(err)
	}
	if confPath = filepath.Dir(selfFile); confPath == "/" {
		confPath = ""
	}

	return conf.NewConfig(
		fmt.Sprintf("/etc/%s/config.yml", project),
		fmt.Sprintf("%s/config.yml", confPath),
	)
}
