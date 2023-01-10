package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/mokiat/gocrane/internal/filesystem"
	"github.com/mokiat/gocrane/internal/project"
)

const ForceBuildPath = "/ffb5c0d8-e6ac-4965-9080-7168f473db57"

func Build(
	ctx context.Context,
	mainDir string,
	buildArgs []string,
	in Queue[ChangeEvent],
	out Queue[BuildEvent],
	rebuildFilter *filesystem.FilterTree,
	restartFilter *filesystem.FilterTree,
	bootstrapEvent *BuildEvent,
) func() error {

	// Create a temporary directory to store binaries.
	tempDir, err := os.MkdirTemp("", "gocrane-*")
	if err != nil {
		return func() error {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
	}
	defer func() {
		os.RemoveAll(tempDir)
	}()

	builder := project.NewBuilder(mainDir, buildArgs)

	return func() error {
		var lastBinary string
		if bootstrapEvent != nil {
			lastBinary = bootstrapEvent.Path
			out.Push(ctx, *bootstrapEvent)
		}

		var changeEvent ChangeEvent
		for in.Pop(ctx, &changeEvent) {
			shouldBuild := isAnyAccepted(rebuildFilter, changeEvent.Paths) || isAnyForceRebuild(changeEvent.Paths)
			shouldRestart := isAnyAccepted(restartFilter, changeEvent.Paths)

			// Skip this change event. The changed files are not of relevance.
			if !shouldBuild && !shouldRestart {
				continue
			}

			// If a restart is requested but there isn't a binary yet, then
			// trigger a build.
			if shouldRestart && (lastBinary == "") {
				shouldBuild = true
			}

			// If just a restart is required, then produce a fake build event
			// based on the last binary.
			if !shouldBuild && shouldRestart {
				out.Push(ctx, BuildEvent{
					Path: lastBinary,
				})
				continue
			}

			log.Printf("Building...")
			path := filepath.Join(tempDir, fmt.Sprintf("executable-%s", uuid.NewString()))
			if err := builder.Build(ctx, path); err != nil {
				log.Printf("Build failure: %s", err)
				continue
			}

			log.Printf("Build was successful.")
			lastBinary = path
			out.Push(ctx, BuildEvent{
				Path: path,
			})
		}

		return nil
	}
}

func isAnyAccepted(filter *filesystem.FilterTree, paths []string) bool {
	for _, path := range paths {
		if filter.IsAccepted(path) {
			return true
		}
	}
	return false
}

func isAnyForceRebuild(paths []string) bool {
	for _, path := range paths {
		if path == ForceBuildPath {
			return true
		}
	}
	return false
}
