package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mokiat/gocrane/internal/crane"
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
				Aliases: []string{"v"},
				Usage:   "verbose logging",
				Value:   false,
			},
			&cli.StringSliceFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "folder(s) to watch for changes",
				Value:   cli.NewStringSlice("./"),
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"e"},
				Usage:   "folders to exclude from watching.",
			},
			&cli.StringFlag{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "directory to build and run",
				Value:   "./",
			},
			&cli.StringFlag{
				Name:    "cache",
				Aliases: []string{"c"},
				Usage:   "prebuilt executable to use initially",
			},
		},
		Action: func(c *cli.Context) error {
			return crane.Run(c.Context, crane.Settings{
				Verbose:      c.Bool("verbose"),
				IncluedPaths: c.StringSlice("path"),
				ExcludePaths: c.StringSlice("exclude"),
				RunDir:       c.String("run"),
				CachedBuild:  c.String("cache"),
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
