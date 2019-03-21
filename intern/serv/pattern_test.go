// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package serv

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Must retrieve queue name from request parameters
func TestPattern_Apply(t *testing.T) {

	samples := []struct {
		pattern string
		query   string
		cookie  string
		expect  string
	}{
		{
			pattern: "",
			query:   "",
			cookie:  "",
			expect:  "",
		},
		{
			pattern: "${unknown:foo}",
			query:   "",
			cookie:  "",
			expect:  "",
		},
		{
			pattern: "${query:oid}",
			query:   "",
			cookie:  "",
			expect:  "",
		},
		{
			pattern: "${query:oid}",
			query:   "oid=",
			cookie:  "",
			expect:  "",
		},
		{
			pattern: "${query:oid}",
			query:   "oid=123",
			cookie:  "",
			expect:  "123",
		},
		{
			pattern: " ${ query : oid } ",
			query:   "oid=123",
			cookie:  "",
			expect:  "123",
		},
		{
			pattern: " x${ query\n : oid }x ",
			query:   "oid=123",
			cookie:  "",
			expect:  "x123x",
		},
		{
			pattern: "${cookie:sid}",
			query:   "",
			cookie:  "",
			expect:  "",
		},
		{
			pattern: "${cookie:sid}",
			query:   "",
			cookie:  "sid=",
			expect:  "",
		},
		{
			pattern: "${cookie:sid}",
			query:   "",
			cookie:  "sid=123",
			expect:  "123",
		},
		{
			pattern: " ${ cookie : sid } ",
			query:   "",
			cookie:  "sid=123",
			expect:  "123",
		},
		{
			pattern: " x${ cookie\n : sid }x ",
			query:   "",
			cookie:  "sid=123",
			expect:  "x123x",
		},
		{
			pattern: "${query:oid}-${cookie:sid}",
			query:   "",
			cookie:  "",
			expect:  "-",
		},
		{
			pattern: "${query:oid}-${cookie:sid}",
			query:   "oid=",
			cookie:  "sid=",
			expect:  "-",
		},
		{
			pattern: "${query:oid}-${cookie:sid}",
			query:   "oid=ABC",
			cookie:  "sid=123",
			expect:  "ABC-123",
		},
		{
			pattern: " ${ query : oid } ${ cookie : sid } ",
			query:   "oid=ABC",
			cookie:  "sid=123",
			expect:  "ABC123",
		},
		{
			pattern: "x${query :oid\n} x${ cookie\n : sid }x ",
			query:   "oid=ABC",
			cookie:  "sid=123",
			expect:  "xABCx123x",
		},
	}

	for i, sample := range samples {
		pattern := NewPattern(sample.pattern)

		req := &http.Request{
			Header: http.Header{"Cookie": {sample.cookie}},
			URL:    &url.URL{RawQuery: sample.query},
		}

		result, _ := pattern.Apply(req)

		if result != sample.expect {
			t.Errorf("(i:%d) expected %#v, got %#v", i, sample.expect, result)
		}
	}
}

func TestPattern_QueueName(t *testing.T) {
	var req *http.Request

	pattern := NewPattern("${query:id}")

	req = &http.Request{URL: &url.URL{RawQuery: ""}}
	if _, err := pattern.Apply(req); err == nil {
		t.Errorf("unexpected result")
	}

	req = &http.Request{URL: &url.URL{RawQuery: "id=" + strings.Repeat("X", 255)}}
	if _, err := pattern.Apply(req); err != nil {
		t.Errorf("unexpected result")
	}

	req = &http.Request{URL: &url.URL{RawQuery: "id=" + strings.Repeat("X", 256)}}
	if _, err := pattern.Apply(req); err == nil {
		t.Errorf("unexpected result")
	}

	req = &http.Request{URL: &url.URL{RawQuery: "id=aMq.foo"}}
	if _, err := pattern.Apply(req); err == nil {
		t.Errorf("unexpected result")
	}

}
