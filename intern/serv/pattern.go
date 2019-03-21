// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package serv

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type (
	// Pattern ...
	Pattern interface {
		Apply(*http.Request) (string, error)
	}
	pattern struct {
		pattern string
		keyVal  *regexp.Regexp
	}
)

var (
	wsReplacer = strings.NewReplacer("\n", "", " ", "")
	tupleRegEx = regexp.MustCompile(`\${([^:]+):([^}]+)}`)

	errNoParams  = errors.New("request parameter(s) missing")
	errQueueName = errors.New("invalid queue name")
)

// NewPattern ...
func NewPattern(s string) Pattern {
	return &pattern{
		pattern: wsReplacer.Replace(s),
		keyVal:  tupleRegEx,
	}
}

// Apply ...
func (p *pattern) Apply(r *http.Request) (string, error) {
	var err error
	var queue string

	repl := func(m string) string {
		var cat, key, val string

		s := p.keyVal.ReplaceAllString(m, "$1 $2")
		_, _ = fmt.Sscanf(s, "%s %s", &cat, &key)

		val, err = resolve(r, cat, key)
		return val
	}

	if queue = p.keyVal.ReplaceAllStringFunc(p.pattern, repl); err != nil {
		return queue, err
	}
	if strings.HasPrefix(strings.ToLower(queue), "amq.") {
		return queue, errQueueName
	}
	if len(queue) == 0 || len(queue) > 255 {
		return queue, errQueueName
	}

	return queue, err
}

func resolve(r *http.Request, cat string, key string) (string, error) {
	switch cat {
	case "cookie":
		return cookieValue(r, key)
	case "query":
		return queryValue(r, key)
	default:
		return "", errNoParams
	}
}

func cookieValue(r *http.Request, key string) (string, error) {
	cookie, err := r.Cookie(key)
	if err != nil {
		return "", err
	}
	if cookie.Value != "" {
		return cookie.Value, nil
	}
	return "", errNoParams
}

func queryValue(r *http.Request, key string) (string, error) {
	if val := r.URL.Query().Get(key); val != "" {
		return val, nil
	}
	return "", errNoParams
}
