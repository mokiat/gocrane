package command

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/mokiat/gocrane/internal/command/flag"
	"github.com/mokiat/gocrane/internal/events"
	"github.com/mokiat/gocrane/internal/location"
	"github.com/mokiat/gocrane/internal/pipeline"
	"github.com/mokiat/gocrane/internal/project"
)

func Run() *cli.Command {
	var cfg runConfig
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newIncludesFlag(&cfg.Includes),
			newExcludesFlag(&cfg.Excludes),
			newSourcesFlag(&cfg.Sources),
			newResourcesFlag(&cfg.Resources),
			newMainFlag(&cfg.MainDir),
			newBinaryFlag(&cfg.BinaryFile, false),
			newBuildArgs(&cfg.BuildArgs),
			newRunArgs(&cfg.RunArgs),
			newBatchDurationFlag(&cfg.BatchDuration),
			newShutdownTimeoutFlag(&cfg.ShutdownTimeout),
		},
		Action: func(c *cli.Context) error {
			return run(c.Context, cfg)
		},
	}
}

type runConfig struct {
	Verbose         bool
	Includes        cli.StringSlice
	Excludes        cli.StringSlice
	Sources         cli.StringSlice
	Resources       cli.StringSlice
	MainDir         string
	BinaryFile      string
	BuildArgs       flag.ShlexStringSlice
	RunArgs         flag.ShlexStringSlice
	BatchDuration   time.Duration
	ShutdownTimeout time.Duration
}

func run(ctx context.Context, cfg runConfig) error {
	log.Println("analyzing project...")
	layout := project.Explore(
		cfg.Includes.Value(),
		cfg.Excludes.Value(),
		cfg.Sources.Value(),
		cfg.Resources.Value(),
	)
	log.Println("project successfully analyzed...")
	if cfg.Verbose {
		layout.PrintToLog()
	}

	var fakeChangeEvent *events.Change
	var fakeBuildEvent *events.Build
	if cfg.BinaryFile != "" {
		log.Println("reading stored digest...")
		digestFile := fmt.Sprintf("%s.dig", cfg.BinaryFile)
		storedDigest, err := project.ReadDigest(digestFile)
		if err != nil {
			return fmt.Errorf("failed to read digest: %w", err)
		}
		log.Println("calculating digest...")
		digest, err := layout.Digest()
		if err != nil {
			return fmt.Errorf("failed to calculate digest: %w", err)
		}
		if storedDigest == digest {
			log.Println("digests match, using binary")
			fakeBuildEvent = &events.Build{
				Path: cfg.BinaryFile,
			}
		} else {
			log.Println("digest mismatch, building from scratch")
			fakeChangeEvent = &events.Change{}
		}
	} else {
		fakeChangeEvent = &events.Change{}
	}

	log.Println("starting pipeline...")
	changeEventQueue := make(events.ChangeQueue, 1024)
	batchChangeEventQueue := make(events.ChangeQueue)
	buildEventQueue := make(events.BuildQueue)

	group, groupCtx := errgroup.WithContext(ctx)

	// Watch for filesystem changes.
	group.Go(pipeline.Watch(
		groupCtx,
		cfg.Verbose,
		layout.WatchDirs,
		location.AndFilter(layout.IncludeFilter, location.NotFilter(layout.ExcludeFilter)),
		changeEventQueue,
		fakeChangeEvent,
	))

	// Accumulate change events and flush them as a single change event
	// once there has been a sufficient period of inactivity.
	// This avoids triggering multiple builds during the continuous change
	// of many files (e.g. git clone / git checkout).
	group.Go(pipeline.Batch(
		groupCtx,
		changeEventQueue,
		batchChangeEventQueue,
		cfg.BatchDuration,
	))

	// Build executable on new batch changes.
	group.Go(pipeline.Build(
		groupCtx,
		cfg.MainDir,
		cfg.BuildArgs.Value(),
		batchChangeEventQueue,
		buildEventQueue,
		location.AndFilter(layout.IncludeFilter, layout.SourcesFilter),
		location.AndFilter(layout.IncludeFilter, layout.ResourcesFilter),
		fakeBuildEvent,
	))

	// Run new executables when built.
	group.Go(pipeline.Run(
		groupCtx,
		cfg.RunArgs.Value(),
		buildEventQueue,
		cfg.ShutdownTimeout,
	))

	if err := group.Wait(); err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	log.Println("stopped.")
	return nil
}
