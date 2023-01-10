package command

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/mokiat/gocrane/internal/command/flag"
	"github.com/mokiat/gocrane/internal/pipeline"
	"github.com/mokiat/gocrane/internal/project"
)

func Run() *cli.Command {
	var cfg runConfig
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newDirFlag(&cfg.Dirs),
			newDirExcludeFlag(&cfg.ExcludeDirs),
			newSourceFlag(&cfg.Sources),
			newSourceExcludeFlag(&cfg.ExcludeSources),
			newResourceFlag(&cfg.Resources),
			newResourceExcludeFlag(&cfg.ExcludeResources),
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
	Verbose          bool
	Dirs             cli.StringSlice
	ExcludeDirs      cli.StringSlice
	Sources          cli.StringSlice
	ExcludeSources   cli.StringSlice
	Resources        cli.StringSlice
	ExcludeResources cli.StringSlice
	MainDir          string
	BinaryFile       string
	BuildArgs        flag.ShlexStringSlice
	RunArgs          flag.ShlexStringSlice
	BatchDuration    time.Duration
	ShutdownTimeout  time.Duration
}

func run(ctx context.Context, cfg runConfig) error {
	log.Println("Preparing filtering...")
	watchFilter, err := buildFilterTree(cfg.Dirs.Value(), cfg.ExcludeDirs.Value())
	if err != nil {
		return fmt.Errorf("problem with dir rules: %w", err)
	}
	sourceFilter, err := buildFilterTree(cfg.Sources.Value(), cfg.ExcludeSources.Value())
	if err != nil {
		return fmt.Errorf("problem with source rules: %w", err)
	}
	resourceFilter, err := buildFilterTree(cfg.Resources.Value(), cfg.ExcludeResources.Value())
	if err != nil {
		return fmt.Errorf("problem with resource rules: %w", err)
	}
	rootDirs := watchFilter.RootPaths()

	var summary *project.Summary
	if cfg.Verbose || cfg.BinaryFile != "" {
		log.Println("Analyzing project...")
		summary = project.Analyze(rootDirs, watchFilter, sourceFilter, resourceFilter)
	}
	if cfg.Verbose {
		printSummary(summary)
	}

	var (
		fakeChangeEvent *pipeline.ChangeEvent
		fakeBuildEvent  *pipeline.BuildEvent
	)
	if cfg.BinaryFile != "" {
		log.Println("Reading stored digest...")
		digestFile := fmt.Sprintf("%s.dig", cfg.BinaryFile)
		storedDigest, err := project.ReadDigest(digestFile)
		if err != nil {
			return fmt.Errorf("failed to read digest: %w", err)
		}

		log.Println("Calculating current digest...")
		digest, err := sourceDigest(summary)
		if err != nil {
			return fmt.Errorf("failed to calculate digest: %w", err)
		}

		log.Println("Comparing stored and current digests...")
		if storedDigest == digest {
			log.Println("\t Digest match, will use existing binary.")
			fakeBuildEvent = &pipeline.BuildEvent{
				Path: cfg.BinaryFile,
			}
		} else {
			log.Printf("\t Digest mismatch (%s != %s), will build from scratch.", digest, storedDigest)
			fakeChangeEvent = &pipeline.ChangeEvent{
				Paths: []string{pipeline.ForceBuildPath},
			}
		}
	} else {
		fakeChangeEvent = &pipeline.ChangeEvent{
			Paths: []string{pipeline.ForceBuildPath},
		}
	}

	log.Println("Running pipeline...")
	changeEventQueue := make(pipeline.Queue[pipeline.ChangeEvent], 1024)
	batchChangeEventQueue := make(pipeline.Queue[pipeline.ChangeEvent])
	buildEventQueue := make(pipeline.Queue[pipeline.BuildEvent])

	group, groupCtx := errgroup.WithContext(ctx)

	// Watch for filesystem changes.
	group.Go(pipeline.Watch(
		groupCtx,
		cfg.Verbose,
		rootDirs,
		watchFilter,
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
		sourceFilter,
		resourceFilter,
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
		return fmt.Errorf("pipeline error: %w", err)
	}

	log.Println("Pipeline stopped.")
	return nil
}
