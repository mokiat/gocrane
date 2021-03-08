package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"

	"github.com/mokiat/gocrane/internal/location"
	"github.com/mokiat/gocrane/internal/project"
)

func Build(
	ctx context.Context,
	mainDir string,
	buildArgs []string,
	in ChangeEventQueue,
	out BuildEventQueue,
	rebuildFilter location.Filter,
	restartFilter location.Filter,
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
			shouldBuild := location.MatchAny(rebuildFilter, changeEvent.Paths)
			shouldRestart := location.MatchAny(restartFilter, changeEvent.Paths)

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

			log.Printf("building...")
			path := filepath.Join(tempDir, fmt.Sprintf("executable-%s", uuid.NewV4()))
			if err := builder.Build(ctx, path); err != nil {
				log.Printf("build failure: %s", err)
				continue
			}

			log.Printf("build was successful.")
			lastBinary = path
			out.Push(ctx, BuildEvent{
				Path: path,
			})
		}

		return nil
	}
}
