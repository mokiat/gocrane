package command

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/flag"
	"github.com/mokiat/gocrane/internal/project"
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

func newSourcesFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:        "source",
		Usage:       "folder(s) and/or file(s) that are required for building the application",
		Aliases:     []string{"src"},
		EnvVars:     []string{"GOCRANE_SOURCES"},
		Value:       cli.NewStringSlice("./"),
		Destination: target,
	}
}

func newResourcesFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:        "resource",
		Usage:       "folder(s) and/or file(s) that are required for running the application",
		Aliases:     []string{"res"},
		EnvVars:     []string{"GOCRANE_RESOURCES"},
		Destination: target,
	}
}

func newIncludesFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:        "include",
		Usage:       "folder(s) and/or file(s) that are of interest for building or running the application",
		Aliases:     []string{"in"},
		EnvVars:     []string{"GOCRANE_INCLUDES"},
		Value:       cli.NewStringSlice(project.Glob("*.go")),
		Destination: target,
	}
}

func newExcludesFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:        "exclude",
		Usage:       "folder(s) and/or file(s) that are not of interest for building or running the application",
		Aliases:     []string{"ex"},
		EnvVars:     []string{"GOCRANE_EXCLUDES"},
		Destination: target,
	}
}

func newMainFlag(target *string) cli.Flag {
	return &cli.StringFlag{
		Name:        "main",
		Usage:       "directory that contains the main package to build",
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
		Aliases:     []string{"bin"},
		EnvVars:     []string{"GOCRANE_BINARY"},
		Destination: target,
	}
}

func newDigestFlag(target *string) cli.Flag {
	return &cli.StringFlag{
		Name:        "digest",
		Usage:       "file that will be used to track the state of sources when running cached applications",
		Aliases:     []string{"dig"},
		EnvVars:     []string{"GOCRANE_DIGEST"},
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

func newNoDefaultExcludes(target *bool) cli.Flag {
	return &cli.BoolFlag{
		Name:        "--no-default-excludes",
		Usage:       "don't use exclude presets",
		EnvVars:     []string{"GOCRANE_NO_DEFAULT_EXCLUDES"},
		Value:       false,
		Destination: target,
	}
}

func newNoDefaultResources(target *bool) cli.Flag {
	return &cli.BoolFlag{
		Name:        "--no-default-resources",
		Usage:       "don't use resource presets",
		EnvVars:     []string{"GOCRANE_NO_DEFAULT_RESOURCES"},
		Value:       false,
		Destination: target,
	}
}
