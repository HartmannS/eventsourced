// MIT license · Daniel T. Gorski · dtg [at] lengo [dot] org · 03/2019

package app

import (
	"fmt"
)

type (
	// Version ...
	Version interface {
		Short() string
		Long() string
		Full() string
	}
	version struct {
		project string
		version string
		commit  string
		status  string
	}
)

// NewVersion ...
func NewVersion(p, v, c, s string) Version {
	return &version{project: p, version: v, commit: c, status: s}
}

// Short ...
func (v *version) Short() string {
	return v.version
}

// Long ...
func (v *version) Long() string {
	if v.status == "" {
		return fmt.Sprintf("%s (%s)", v.version, v.commit)
	}
	return fmt.Sprintf("%s (%s-%s)", v.version, v.commit, v.status)

}

// Full ...
func (v *version) Full() string {
	return fmt.Sprintf("%s %s", v.project, v.Long())
}
