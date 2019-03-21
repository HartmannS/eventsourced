// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package app

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"eventsourced/intern/conf"
)

func TestGetOpt_NewGetOpt(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(ioutil.Discard)

	if flag.Parsed() != false {
		t.Errorf("expected flags to be not parsed")
	}

	_ = NewGetOpt()

	if flag.Parsed() != true {
		t.Errorf("expected flags to be parsed")
	}
}

func TestGetOpt_Complement(t *testing.T) {
	True, False := true, false

	out := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = out }()

	state := NewState(NewVersion("", "", "", ""), conf.NewConfig())

	o := &getOpt{version: &True}
	if o.Complement(state) != false {
		t.Errorf("expected false, got true")
	}

	o = &getOpt{version: &False, confdump: &True}
	if o.Complement(state) != false {
		t.Errorf("expected false, got true")
	}

	o = &getOpt{version: &False, confdump: &False}
	if o.Complement(state) != true {
		t.Errorf("expected true, got false")
	}
}
