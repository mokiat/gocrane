package crane

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/mokiat/gocrane/internal/change"
	"github.com/mokiat/gocrane/internal/events"
	"github.com/mokiat/gocrane/internal/project"
	"golang.org/x/sync/errgroup"
)

type Settings struct {
	Verbose         bool
	IncluedPaths    []string
	ExcludePaths    []string
	ExcludeGlobs    []string
	RunDir          string
	CachedBuild     string
	ShutdownTimeout time.Duration
}

func Run(ctx context.Context, settings Settings) error {
	log.Println("starting...")

	for _, glob := range settings.ExcludeGlobs {
		if _, err := filepath.Match(glob, ""); err != nil {
			return fmt.Errorf("invalid glob exclude pattern %q: %w", glob, err)
		}
	}

	defer log.Println("stopped.")

	group, groupCtx := errgroup.WithContext(ctx)

	// These are the queues with which we progress changes
	// through the pipeline.
	changeEventQueue := make(events.ChangeQueue, 1024)
	batchChangeEventQueue := make(events.ChangeQueue)
	buildEventQueue := make(events.BuildQueue)

	watcher := change.NewWatcher(settings.IncluedPaths, settings.ExcludePaths, settings.ExcludeGlobs, settings.Verbose)
	batcher := change.NewBatcher(time.Second)
	builder, err := project.NewBuilder(settings.RunDir)
	if err != nil {
		return fmt.Errorf("failed to create builder: %w", err)
	}
	defer builder.Cleanup()
	runner := project.NewRunner()

	// Trigger an initial build by faking a change if we don't have
	// a cached executable to use.
	if settings.CachedBuild == "" {
		go func() {
			changeEventQueue <- events.Change{}
		}()
	}

	// Trigger an initial run by faking a build change if we have
	// a cached executable specified.
	if settings.CachedBuild != "" {
		go func() {
			buildEventQueue <- events.Build{
				Path: settings.CachedBuild,
			}
		}()
	}

	// Watch for filesystem changes.
	group.Go(func() error {
		return watcher.Run(groupCtx, changeEventQueue)
	})

	// Accumulate change events and flush them as a single change event
	// once there has been a sufficient period of inactivity.
	// This avoids triggering multiple builds during the continuous change
	// of many files (e.g. git clone / git checkout).
	group.Go(func() error {
		defer close(batchChangeEventQueue)
		return batcher.Run(groupCtx, changeEventQueue, batchChangeEventQueue)
	})

	// Build executable on new batch changes.
	group.Go(createBuildFlow(
		groupCtx,
		builder,
		batchChangeEventQueue,
		buildEventQueue,
	))

	// Run new executables when built.
	group.Go(createRunFlow(
		groupCtx,
		runner,
		buildEventQueue,
		settings.ShutdownTimeout,
	))

	return group.Wait()
}

func createBuildFlow(
	ctx context.Context,
	builder *project.Builder,
	batchChangeEventQueue events.ChangeQueue,
	buildEventQueue events.BuildQueue,
) func() error {

	return func() error {
		for batchChangeEventQueue.Pop(ctx, &events.Change{}) {
			log.Printf("building...")

			path, err := builder.Build(ctx)
			if err != nil {
				log.Printf("build failure: %s", err)
				continue
			}

			log.Printf("build was successful.")
			buildEventQueue.Push(ctx, events.Build{
				Path: path,
			})
		}
		return nil
	}
}

func createRunFlow(
	ctx context.Context,
	runner *project.Runner,
	buildEventQueue events.BuildQueue,
	shutdownTimeout time.Duration,
) func() error {

	return func() error {
		var runningProcess *project.Process

		startProcess := func(path string) error {
			if runningProcess != nil {
				return fmt.Errorf("there is already a running process")
			}
			log.Printf("starting new process...")
			process, err := runner.Run(ctx, path)
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
			log.Printf("stopping running process...")
			shutdownCtx, shutdownFunc := context.WithTimeout(ctx, shutdownTimeout)
			defer shutdownFunc()
			if err := runningProcess.Stop(shutdownCtx); err != nil {
				return fmt.Errorf("failed to stop process: %w", err)
			}
			log.Printf("successfully stopped running process.")
			runningProcess = nil
			return nil
		}

		var buildEvent events.Build
		for buildEventQueue.Pop(ctx, &buildEvent) {
			if err := stopProcess(); err != nil {
				return err
			}
			if err := startProcess(buildEvent.Path); err != nil {
				return err
			}
		}
		return stopProcess()
	}
}
