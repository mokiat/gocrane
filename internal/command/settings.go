package command

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/crane"
	"github.com/mokiat/gocrane/internal/flag"
)

func newSettings(c *cli.Context) (*crane.Settings, error) {
	return &crane.Settings{
		Verbose:         c.Bool("verbose"),
		Sources:         c.StringSlice("source"),
		Resources:       c.StringSlice("resource"),
		Excludes:        c.StringSlice("exclude"),
		MainDir:         c.String("main"),
		BinaryFile:      c.String("binary"),
		DigestFile:      c.String("digest"),
		BuildArgs:       flag.ShlexStrings(c.Generic("build-arg")),
		RunArgs:         flag.ShlexStrings(c.Generic("run-arg")),
		ShutdownTimeout: c.Duration("shutdown-timeout"),
	}, nil
}

func newBuildSettings(c *cli.Context) (*crane.BuildSettings, error) {
	settings, err := newSettings(c)
	if err != nil {
		return nil, err
	}
	if settings.MainDir == "" {
		return nil, fmt.Errorf("main dir needs to be specified")
	}
	if settings.BinaryFile == "" {
		return nil, fmt.Errorf("binary file needs to be specified")
	}
	return &crane.BuildSettings{
		Settings: *settings,
	}, nil
}

func newRunSettings(c *cli.Context) (*crane.RunSettings, error) {
	settings, err := newSettings(c)
	if err != nil {
		return nil, err
	}
	return &crane.RunSettings{
		Settings: *settings,
	}, nil
}
