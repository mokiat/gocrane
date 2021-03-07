package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"

	"github.com/mokiat/gocrane/internal/events"
	"github.com/mokiat/gocrane/internal/location"
)

func Watch(
	ctx context.Context,
	verbose bool,
	dirs []string,
	watchFilter location.Filter,
	out events.ChangeQueue,
	bootstrapEvent *events.Change,

) func() error {
	isEventType := func(event fsnotify.Event, eType fsnotify.Op) bool {
		return event.Op&eType == eType
	}

	return func() error {
		if bootstrapEvent != nil {
			out.Push(ctx, *bootstrapEvent)
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("failed to create filesystem watcher: %w", err)
		}
		defer watcher.Close()

		watchDir := func(path string) {
			if err := watcher.Add(path); err != nil {
				log.Printf("failed to watch %q: %v", path, err)
			} else {
				if verbose {
					log.Printf("watching %q", path)
				}
			}
		}

		processFSEvent := func(event fsnotify.Event) {
			if verbose {
				log.Printf("filesystem watch event: %s", event)
			}
			path, err := filepath.Abs(event.Name)
			if err != nil {
				log.Printf("failed to convert path to absolute %q: %v", event.Name, err)
				return
			}
			if !watchFilter.Match(path) {
				if verbose {
					log.Printf("skipping excluded path %q from watching", path)
				}
				return
			}
			stat, err := os.Stat(path)
			if err != nil {
				log.Printf("failed to stat file %q: %v", event.Name, err)
				return
			}
			if stat.IsDir() {
				if isEventType(event, fsnotify.Create) {
					watchDir(path)
				}
			} else {
				if !isEventType(event, fsnotify.Chmod) {
					out.Push(ctx, events.Change{
						Paths: []string{path},
					})
				}
			}
		}

		for _, dir := range dirs {
			watchDir(dir)
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case event := <-watcher.Events:
				processFSEvent(event)
			case err := <-watcher.Errors:
				log.Printf("filesystem watcher error: %v", err)
			}
		}
	}
}
