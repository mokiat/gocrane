// Package flag provides types to be used as command-line flag arguments with
// the github.com/urfave/cli package.
package flag

import (
	"strings"

	"github.com/google/shlex"
	"github.com/urfave/cli/v3"
)

type ShlexStringSlice struct {
	slice []string
}

var _ cli.Value = (*ShlexStringSlice)(nil)

func (s *ShlexStringSlice) Items() []string {
	return s.slice
}

func (s *ShlexStringSlice) Set(value string) error {
	args, err := shlex.Split(value)
	if err != nil {
		return err
	}
	s.slice = args
	return nil
}

func (s *ShlexStringSlice) Get() any {
	return s.slice
}

func (s *ShlexStringSlice) String() string {
	return strings.Join(s.slice, " ")
}
