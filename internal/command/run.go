package command

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/mokiat/gocrane/internal/change"
	"github.com/mokiat/gocrane/internal/command/flag"
	"github.com/mokiat/gocrane/internal/events"
	"github.com/mokiat/gocrane/internal/project"
)

func Run() *cli.Command {
	var cfg runConfig
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			newVerboseFlag(&cfg.Verbose),
			newSourcesFlag(&cfg.Sources),
			newResourcesFlag(&cfg.Resources),
			newIncludesFlag(&cfg.Includes),
			newExcludesFlag(&cfg.Excludes),
			newMainFlag(&cfg.MainDir),
			newBinaryFlag(&cfg.BinaryFile, false),
			newDigestFlag(&cfg.DigestFile),
			newBuildArgs(&cfg.BuildArgs),
			newRunArgs(&cfg.RunArgs),
			newBatchDurationFlag(&cfg.BatchDuration),
			newShutdownTimeoutFlag(&cfg.ShutdownTimeout),
			newNoDefaultExcludes(&cfg.NoDefaultExcludes),
			newNoDefaultResources(&cfg.NoDefaultResources),
		},
		Action: func(c *cli.Context) error {
			return run(c.Context, cfg)
		},
	}
}

type runConfig struct {
	Verbose            bool
	Sources            cli.StringSlice
	Resources          cli.StringSlice
	Includes           cli.StringSlice
	Excludes           cli.StringSlice
	MainDir            string
	BinaryFile         string
	DigestFile         string
	BuildArgs          flag.ShlexStringSlice
	RunArgs            flag.ShlexStringSlice
	BatchDuration      time.Duration
	ShutdownTimeout    time.Duration
	NoDefaultExcludes  bool
	NoDefaultResources bool
}

func run(ctx context.Context, cfg runConfig) error {
	log.Println("analyzing project...")
	excludes := cfg.Excludes.Value()
	if !cfg.NoDefaultExcludes {
		excludes = addDefaultExcludes(excludes)
	}
	resources := cfg.Resources.Value()
	if !cfg.NoDefaultResources {
		resources = addDefaultResources(resources)
	}
	layout, err := project.Explore(
		cfg.Sources.Value(),
		project.NewFilter(resources),
		project.NewFilter(cfg.Includes.Value()),
		project.NewFilter(excludes),
	)
	if err != nil {
		return fmt.Errorf("failed to explore project: %w", err)
	}
	if cfg.Verbose {
		logLayout(layout)
	}

	var fakeChangeEvent *events.Change
	var fakeBuildEvent *events.Build
	if cfg.BinaryFile != "" && cfg.DigestFile != "" {
		log.Println("reading stored digest...")
		storedDigest, err := readDigest(cfg.DigestFile)
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

	log.Println("creating temp build directory...")
	tempDir, err := os.MkdirTemp("", "gocrane-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	group, groupCtx := errgroup.WithContext(ctx)

	// Watch for filesystem changes.
	changeEventQueue := make(events.ChangeQueue, 1024)
	group.Go(func() error {
		if fakeChangeEvent != nil {
			changeEventQueue <- *fakeChangeEvent
		}
		dirs := project.CombineFileSets(layout.SourceDirs, layout.ResourceDirs)
		watcher := change.NewWatcher(cfg.Verbose, dirs, layout.ExcludeFilter)
		return watcher.Run(groupCtx, changeEventQueue)
	})

	// Accumulate change events and flush them as a single change event
	// once there has been a sufficient period of inactivity.
	// This avoids triggering multiple builds during the continuous change
	// of many files (e.g. git clone / git checkout).
	batchChangeEventQueue := make(events.ChangeQueue)
	group.Go(func() error {
		batcher := change.NewBatcher(cfg.BatchDuration)
		return batcher.Run(groupCtx, changeEventQueue, batchChangeEventQueue)
	})

	// Build executable on new batch changes.
	buildEventQueue := make(events.BuildQueue)
	group.Go(func() error {
		var lastBinary string
		if fakeBuildEvent != nil {
			lastBinary = fakeBuildEvent.Path
			buildEventQueue <- *fakeBuildEvent
		}
		var changeEvent events.Change
		for batchChangeEventQueue.Pop(groupCtx, &changeEvent) {
			// changes to resources should only yield a restart
			if layout.ResourceFilter.MatchAll(changeEvent.Paths...) && (lastBinary != "") {
				buildEventQueue <- events.Build{
					Path: lastBinary,
				}
				continue
			}

			log.Printf("building...")
			path := filepath.Join(tempDir, fmt.Sprintf("executable-%s", uuid.NewV4()))
			builder := project.NewBuilder(cfg.MainDir, cfg.BuildArgs.Value(), path)
			if err := builder.Build(groupCtx); err != nil {
				log.Printf("build failure: %s", err)
				continue
			}

			log.Printf("build was successful.")
			lastBinary = path
			buildEventQueue.Push(groupCtx, events.Build{
				Path: path,
			})
		}
		return nil
	})

	// Run new executables when built.
	group.Go(func() error {
		var runningProcess *project.Process
		runner := project.NewRunner(cfg.RunArgs.Value())

		startProcess := func(path string) error {
			if runningProcess != nil {
				return fmt.Errorf("there is already a running process")
			}
			log.Printf("starting new process...")
			process, err := runner.Run(context.Background(), path)
			if err != nil {
				return fmt.Errorf("failed to start process: %w", err)
			}
			log.Printf("successfully started new process.")
			runningProcess = process
			return nil
		}

		stopProcess := func() error {
			if runningProcess == nil {
				return nil
			}
			log.Printf("stopping running process (timeout: %s)...", cfg.ShutdownTimeout)
			shutdownCtx, shutdownFunc := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer shutdownFunc()
			if err := runningProcess.Stop(shutdownCtx); err != nil {
				return fmt.Errorf("failed to stop process: %w", err)
			}
			log.Printf("successfully stopped running process.")
			runningProcess = nil
			return nil
		}

		var buildEvent events.Build
		for buildEventQueue.Pop(groupCtx, &buildEvent) {
			if err := stopProcess(); err != nil {
				return err
			}
			if err := startProcess(buildEvent.Path); err != nil {
				return err
			}
		}
		return stopProcess()
	})

	if err := group.Wait(); err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	log.Println("stopped.")
	return nil
}

func readDigest(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("failed to open file %q: %w", file, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read from file %q: %w", file, err)
	}
	return string(data), nil
}
