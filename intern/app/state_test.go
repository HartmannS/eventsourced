// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package app

import (
	"reflect"
	"testing"

	"eventsourced/intern/conf"
)

func TestState(t *testing.T) {
	version := NewVersion("1", "2", "3", "4")
	config := conf.NewConfig("")

	state := NewState(version, config)

	if !reflect.DeepEqual(state.Version(), version) {
		t.Error("expected equality")
	}
	if !reflect.DeepEqual(state.Config(), config) {
		t.Error("expected equality")
	}
}
