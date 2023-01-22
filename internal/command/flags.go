package command

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/command/flag"
	"github.com/mokiat/gocrane/internal/filesystem"
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

func newDirFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:    "dir",
		Usage:   "folder(s) that should be watched",
		Aliases: []string{"d"},
		EnvVars: []string{"GOCRANE_DIRS"},
		Value: cli.NewStringSlice(
			"./",
		),
		Destination: target,
	}
}

func newDirExcludeFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:    "dir-exclude",
		Usage:   "filter(s) for folders that should be excluded from watching",
		Aliases: []string{"de"},
		EnvVars: []string{"GOCRANE_DIR_EXCLUDES"},
		Value: cli.NewStringSlice(
			filesystem.Glob(".git"),
			filesystem.Glob(".github"),
			filesystem.Glob(".gitlab"),
			filesystem.Glob(".vscode"),
		),
		Destination: target,
	}
}

func newSourceFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:    "source",
		Usage:   "filter(s) that indicate which watched files should trigger a build",
		Aliases: []string{"s"},
		EnvVars: []string{"GOCRANE_SOURCES"},
		Value: cli.NewStringSlice(
			filesystem.Glob("*.go"),
			filesystem.Glob("go.mod"),
			filesystem.Glob("go.sum"),
		),
		Destination: target,
	}
}

func newSourceExcludeFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:    "source-exclude",
		Usage:   "filter(s) that indicate which watched files should not trigger a build",
		Aliases: []string{"se"},
		EnvVars: []string{"GOCRANE_SOURCE_EXCLUDES"},
		Value: cli.NewStringSlice(
			filesystem.Glob("*_test.go"),
		),
		Destination: target,
	}
}

func newResourceFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:        "resource",
		Usage:       "filter(s) that indicate which watched files should trigger a restart",
		Aliases:     []string{"r"},
		EnvVars:     []string{"GOCRANE_RESOURCES"},
		Value:       cli.NewStringSlice(),
		Destination: target,
	}
}

func newResourceExcludeFlag(target *cli.StringSlice) cli.Flag {
	return &cli.StringSliceFlag{
		Name:    "resource-exclude",
		Usage:   "filter(s) that indicate which watched files should not trigger a restart",
		Aliases: []string{"re"},
		EnvVars: []string{"GOCRANE_RESOURCE_EXCLUDES"},
		Value: cli.NewStringSlice(
			filesystem.Glob(".gitignore"),
			filesystem.Glob(".gitattributes"),
			filesystem.Glob(".DS_Store"),
			filesystem.Glob("README.md"),
			filesystem.Glob("LICENSE"),
			filesystem.Glob("Dockerfile"),
			filesystem.Glob("docker-compose.yml"),
		),
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
