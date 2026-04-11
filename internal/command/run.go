package command

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/mokiat/gocrane/internal/command/flag"
	"github.com/mokiat/gocrane/internal/pipeline"
	"github.com/mokiat/gocrane/internal/project"
	"github.com/mokiat/gog/opt"
)

func Run() *cli.Command {
	var cfg runConfig
	return &cli.Command{
		Name:        "run",
		Usage:       "run and watch the go application",
		Description: "use this command to run a go application and automatically rebuild and rerun it when source files change",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newPWDFlag(&cfg.PWD),
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
		Action: func(ctx context.Context, _ *cli.Command) error {
			return run(ctx, cfg)
		},
	}
}

type runConfig struct {
	Verbose          bool
	PWD              string
	Dirs             []string
	ExcludeDirs      []string
	Sources          []string
	ExcludeSources   []string
	Resources        []string
	ExcludeResources []string
	MainDir          string
	BinaryFile       string
	BuildArgs        flag.ShlexStringSlice
	RunArgs          flag.ShlexStringSlice
	BatchDuration    time.Duration
	ShutdownTimeout  time.Duration
}

func run(ctx context.Context, cfg runConfig) error {
	log.Println("Preparing filtering...")
	watchFilter, err := buildFilterTree(cfg.Dirs, cfg.ExcludeDirs)
	if err != nil {
		return fmt.Errorf("problem with dir rules: %w", err)
	}
	sourceFilter, err := buildFilterTree(cfg.Sources, cfg.ExcludeSources)
	if err != nil {
		return fmt.Errorf("problem with source rules: %w", err)
	}
	resourceFilter, err := buildFilterTree(cfg.Resources, cfg.ExcludeResources)
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

	var bootstrapEvent pipeline.RestartEvent
	if cfg.BinaryFile != "" {
		log.Println("Reading stored digest...")
		digestFile := fmt.Sprintf("%s.dig", cfg.BinaryFile)
		storedDigest, err := project.OpenDigestFile(digestFile)
		if err != nil {
			return fmt.Errorf("failed to read digest: %w", err)
		}

		log.Println("Calculating current digest...")
		digest, err := calculateDigest(summary)
		if err != nil {
			return fmt.Errorf("failed to calculate digest: %w", err)
		}

		log.Println("Comparing stored and current digests...")
		if storedDigest == digest {
			log.Println("\t Digest match, will use existing binary.")
			bootstrapEvent.ShouldRebuild = false
		} else {
			log.Printf("\t Digest mismatch (%s != %s), will build from scratch.", digest, storedDigest)
			bootstrapEvent.ShouldRebuild = true
		}
	} else {
		bootstrapEvent.ShouldRebuild = true
	}

	// Prepare pipeline nodes.
	watchNode := pipeline.NewWatchNode(
		rootDirs,
		watchFilter.IsAccepted,
		cfg.Verbose,
	)
	decisionNode := pipeline.NewDecisionNode(
		sourceFilter.IsAccepted,
		resourceFilter.IsAccepted,
	)
	batcherNode := pipeline.NewBatcherNode(
		cfg.BatchDuration,
	)
	lifecycleNode := pipeline.NewLifecycleNode(
		project.NewBuilder(cfg.MainDir, cfg.BuildArgs.Items()),
		project.NewRunner(cfg.PWD, cfg.RunArgs.Items()),
		opt.Wrap(cfg.BinaryFile, cfg.BinaryFile != ""),
		cfg.ShutdownTimeout,
	)

	// Prepare pipeline events.
	changeEvents := make(pipeline.Queue[pipeline.ChangeEvent], 32)
	restartEvents := make(pipeline.Queue[pipeline.RestartEvent], 32)
	batchedRestartEvents := make(pipeline.Queue[pipeline.RestartEvent], 1)
	batchedRestartEvents <- bootstrapEvent // queue initial build and/or run

	// Run pipeline nodes.
	log.Println("Running pipeline...")
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return watchNode.Run(groupCtx, changeEvents)
	})
	group.Go(func() error {
		return decisionNode.Run(groupCtx, changeEvents, restartEvents)
	})
	group.Go(func() error {
		return batcherNode.Run(groupCtx, restartEvents, batchedRestartEvents)
	})
	group.Go(func() error {
		return lifecycleNode.Run(groupCtx, batchedRestartEvents)
	})
	if err := group.Wait(); err != nil {
		return fmt.Errorf("pipeline error: %w", err)
	}

	log.Println("Pipeline stopped.")
	return nil
}
