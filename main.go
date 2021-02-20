package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mokiat/gocrane/internal/crane"
	"github.com/mokiat/gocrane/internal/flag"
	"github.com/urfave/cli/v2"
)

func main() {
	log.SetPrefix("[gocrane]: ")
	log.SetFlags(log.Ltime | log.Lmsgprefix)

	app := &cli.App{
		Name:  "gocrane",
		Usage: "run go executables in docker environment",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "verbose logging",
				Aliases: []string{"v"},
				EnvVars: []string{"GOCRANE_VERBOSE"},
				Value:   false,
			},
			&cli.StringSliceFlag{
				Name:    "path",
				Usage:   "folder(s) to watch for changes",
				Aliases: []string{"p"},
				EnvVars: []string{"GOCRANE_PATHS"},
				Value:   cli.NewStringSlice("./"),
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Usage:   "folders to exclude from watching",
				Aliases: []string{"e"},
				EnvVars: []string{"GOCRANE_EXCLUDES"},
			},
			&cli.StringSliceFlag{
				Name:    "glob-exclude",
				Usage:   "glob(s) to exclude from watching",
				Aliases: []string{"ge"},
				EnvVars: []string{"GOCRANE_GLOB_EXCLUDES"},
			},
			&cli.StringFlag{
				Name:    "run",
				Usage:   "directory to build and run",
				Aliases: []string{"r"},
				EnvVars: []string{"GOCRANE_RUN"},
				Value:   "./",
			},
			&cli.StringFlag{
				Name:    "cache",
				Usage:   "prebuilt executable to use initially",
				Aliases: []string{"c"},
				EnvVars: []string{"GOCRANE_CACHE"},
			},
			&cli.GenericFlag{
				Name:    "args",
				Usage:   "arguments to use when running the built executable",
				EnvVars: []string{"GOCRANE_ARGS"},
				Value:   &flag.ShlexStringSlice{},
			},
			&cli.DurationFlag{
				Name:    "shutdown-timeout",
				Usage:   "amount of time to wait for program to exit gracefully",
				Value:   5 * time.Second,
				EnvVars: []string{"GOCRANE_SHUTDOWN_TIMEOUT"},
			},
		},
		Action: func(c *cli.Context) error {
			return crane.Run(c.Context, crane.Settings{
				Verbose:         c.Bool("verbose"),
				IncluedPaths:    c.StringSlice("path"),
				ExcludePaths:    c.StringSlice("exclude"),
				ExcludeGlobs:    c.StringSlice("glob-exclude"),
				RunDir:          c.String("run"),
				CachedBuild:     c.String("cache"),
				Args:            flag.ShlexStrings(c.Generic("args")),
				ShutdownTimeout: c.Duration("shutdown-timeout"),
			})
		},
	}

	appCtx, appStop := context.WithCancel(context.Background())
	defer appStop()
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigChan)
		<-sigChan
		appStop()
	}()

	if err := app.RunContext(appCtx, os.Args); err != nil {
		log.Fatalf("crashed due to: %s", err)
	}
}
