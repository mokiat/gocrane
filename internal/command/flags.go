package command

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/command/flag"
	"github.com/mokiat/gocrane/internal/location"
)

func newVerboseFlag(target *bool) cli.Flag {
	return &cli.BoolFlag{
		Name:        "verbose",
		Usage:       "verbose logging",
		Aliases:     []string{"v"},
		EnvVars:     []string{"GOCRANE_VERBOSE"},
		Value:       false,
		Destination: target,
	}
}

func newIncludesFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:        "include",
		Usage:       "folder(s) and/or file(s) that should be considered",
		Aliases:     []string{"i"},
		EnvVars:     []string{"GOCRANE_INCLUDES"},
		Value:       cli.NewStringSlice("./"),
		Destination: target,
	}
}

func newExcludesFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:    "exclude",
		Usage:   "folder(s) and/or file(s) that should be ignored",
		Aliases: []string{"e"},
		EnvVars: []string{"GOCRANE_EXCLUDES"},
		Value: cli.NewStringSlice(
			location.Glob(".git"),
			location.Glob(".gitignore"),
			location.Glob(".gitattributes"),
			location.Glob(".github"),
			location.Glob(".gitlab"),
			location.Glob(".vscode"),
			location.Glob(".DS_Store"),
		),
		Destination: target,
	}
}

func newSourcesFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:    "source",
		Usage:   "filter(s) that indicate which files should trigger a build",
		Aliases: []string{"src"},
		EnvVars: []string{"GOCRANE_SOURCES"},
		Value: cli.NewStringSlice(
			location.Glob("*.go"),
		),
		Destination: target,
	}
}

func newResourcesFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:        "resource",
		Usage:       "filter(s) that indicate which files should trigger a restart",
		Aliases:     []string{"res"},
		EnvVars:     []string{"GOCRANE_RESOURCES"},
		Destination: target,
	}
}

func newMainFlag(target *string) cli.Flag {
	return &cli.StringFlag{
		Name:        "main",
		Usage:       "directory that contains the main package to build",
		Aliases:     []string{"m"},
		EnvVars:     []string{"GOCRANE_MAIN"},
		Value:       "./",
		Destination: target,
	}
}

func newBinaryFlag(target *string, required bool) cli.Flag {
	return &cli.StringFlag{
		Name:        "binary",
		Usage:       "file that will be used to build or run an initial (cached) application",
		Required:    required,
		Aliases:     []string{"b", "bin"},
		EnvVars:     []string{"GOCRANE_BINARY"},
		Destination: target,
	}
}

func newBuildArgs(target *flag.ShlexStringSlice) cli.Flag {
	return &cli.GenericFlag{
		Name:    "build-args",
		Usage:   "arguments to use when building the executable",
		Aliases: []string{"ba"},
		EnvVars: []string{"GOCRANE_BUILD_ARGS"},
		Value:   target,
	}
}

func newRunArgs(target *flag.ShlexStringSlice) cli.Flag {
	return &cli.GenericFlag{
		Name:    "run-args",
		Usage:   "arguments to use when running the built executable",
		Aliases: []string{"ra"},
		EnvVars: []string{"GOCRANE_RUN_ARGS"},
		Value:   target,
	}
}

func newBatchDurationFlag(target *time.Duration) cli.Flag {
	return &cli.DurationFlag{
		Name:        "batch-duration",
		Usage:       "amount of time to accumulate change events before triggering a build",
		Value:       time.Second,
		Aliases:     []string{"bd"},
		EnvVars:     []string{"GOCRANE_BATCH_DURATION"},
		Destination: target,
	}
}

func newShutdownTimeoutFlag(target *time.Duration) cli.Flag {
	return &cli.DurationFlag{
		Name:        "shutdown-timeout",
		Usage:       "amount of time to wait for the application to exit gracefully",
		Value:       5 * time.Second,
		Aliases:     []string{"st"},
		EnvVars:     []string{"GOCRANE_SHUTDOWN_TIMEOUT"},
		Destination: target,
	}
}
