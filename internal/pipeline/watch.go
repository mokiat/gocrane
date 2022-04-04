package pipeline

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/mokiat/gocrane/internal/location"
)

func Watch(
	ctx context.Context,
	verbose bool,
	dirs []string,
	watchFilter location.Filter,
	out Queue[ChangeEvent],
	bootstrapEvent *ChangeEvent,

) func() error {
	isEventType := func(event fsnotify.Event, eType fsnotify.Op) bool {
		return event.Op&eType == eType
	}

	return func() error {
		if bootstrapEvent != nil {
			if !out.Push(ctx, *bootstrapEvent) {
				return nil
			}
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("failed to create filesystem watcher: %w", err)
		}
		defer watcher.Close()

		watchedPaths := make(map[string]struct{})

		watchPath := func(root string) map[string]struct{} {
			result := location.Traverse(root, watchFilter, func(path string, isDir bool) error {
				watchedPaths[path] = struct{}{}
				if !isDir {
					return location.ErrSkip
				}
				if err := watcher.Add(path); err != nil {
					return fmt.Errorf("failed to watch %q: %w", path, err)
				}
				return nil
			})
			for path, err := range result.ErroredPaths {
				log.Printf("failed to watch %q: %v", path, err)
			}
			if verbose {
				for path := range result.VisitedPaths {
					log.Printf("watching %q", path)
				}
			}
			return result.VisitedPaths
		}

		unwatchPath := func(root string) map[string]struct{} {
			result := make(map[string]struct{})
			for p := range watchedPaths {
				if strings.HasPrefix(p, root) {
					result[p] = struct{}{}
					delete(watchedPaths, p)
				}
			}
			return result
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
					log.Printf("skipping excluded path %q from processing", path)
				}
				return
			}

			switch {
			case isEventType(event, fsnotify.Create):
				paths := watchPath(path)
				event := ChangeEvent{}
				for path := range paths {
					event.Paths = append(event.Paths, path)
				}
				out.Push(ctx, event)

			case isEventType(event, fsnotify.Remove):
				paths := unwatchPath(path)
				event := ChangeEvent{}
				for path := range paths {
					event.Paths = append(event.Paths, path)
				}
				out.Push(ctx, event)

			case isEventType(event, fsnotify.Chmod):
				// We do nothing on these, since MacOS produces a lot of them.

			default:
				out.Push(ctx, ChangeEvent{
					Paths: []string{path},
				})
			}
		}

		for _, dir := range dirs {
			watchPath(dir)
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
