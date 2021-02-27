// Package flag provides types to be used as command-line flag arguments with
// the github.com/urfave/cli package.
package flag

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/shlex"
)

var shlexMagicPrefix = fmt.Sprintf("shlex:::%d:::", time.Now().UTC().UnixNano())

type ShlexStringSlice struct {
	slice      []string
	hasBeenSet bool
}

func (s *ShlexStringSlice) Set(value string) error {
	if !s.hasBeenSet {
		s.slice = []string{}
		s.hasBeenSet = true
	}

	// This is some magic flow that is required to get cli to work
	// correctly when the flag is specified multiple times. The
	// cli.StringSlice implementation was used as reference.
	if strings.HasPrefix(value, shlexMagicPrefix) {
		trimmed := strings.TrimPrefix(value, shlexMagicPrefix)
		if err := json.Unmarshal([]byte(trimmed), &s.slice); err != nil {
			return err
		}
		return nil
	}

	args, err := shlex.Split(value)
	if err != nil {
		return err
	}
	s.slice = append(s.slice, args...)
	return nil
}

func (s *ShlexStringSlice) String() string {
	return strings.Join(s.slice, ",")
}

func (s *ShlexStringSlice) Serialize() string {
	jsonBytes, _ := json.Marshal(s.slice)
	return fmt.Sprintf("%s%s", shlexMagicPrefix, string(jsonBytes))
}
