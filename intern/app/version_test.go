// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package app

import (
	"testing"
)

func TestVersion(t *testing.T) {
	var expect, result string
	var version Version

	version = NewVersion("1", "2", "3", "4")

	if expect, result = "2", version.Short(); expect != result {
		t.Errorf("expected %s, got %s", expect, result)
	}
	if expect, result = "2 (3-4)", version.Long(); expect != result {
		t.Errorf("expected %s, got %s", expect, result)
	}
	if expect, result = "1 2 (3-4)", version.Full(); expect != result {
		t.Errorf("expected %s, got %s", expect, result)
	}

	version = NewVersion("1", "2", "3", "")

	if expect, result = "2 (3)", version.Long(); expect != result {
		t.Errorf("expected %s, got %s", expect, result)
	}
	if expect, result = "1 2 (3)", version.Full(); expect != result {
		t.Errorf("expected %s, got %s", expect, result)
	}
}
