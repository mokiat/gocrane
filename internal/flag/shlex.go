// Package flag provides types to be used as command-line flag arguments with
// the github.com/urfave/cli package.
package flag

import (
	"fmt"

	"github.com/google/shlex"
)

type ShlexStringSlice []string

func (s *ShlexStringSlice) Set(value string) error {
	v, err := shlex.Split(value)
	if err != nil {
		return err
	}
	*s = v
	return nil
}

func (s *ShlexStringSlice) String() string {
	return fmt.Sprintf("%v", []string(*s))
}

func ShlexStrings(val interface{}) []string {
	s, ok := val.(*ShlexStringSlice)
	if !ok {
		return nil
	}

	if s == nil {
		return nil
	}

	return []string(*s)
}
