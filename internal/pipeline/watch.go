package pipeline

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/mokiat/gocrane/internal/filesystem"
)

func Watch(
	ctx context.Context,
	verbose bool,
	dirs []string,
	watchFilter *filesystem.FilterTree,
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
			result := make(map[string]struct{})
			filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					log.Printf("Error traversing %q: %v", p, err)
					return filepath.SkipDir
				}
				absPath, err := filesystem.ToAbsolutePath(p)
				if err != nil {
					log.Printf("Error converting path %q to absolute: %v", p, err)
					return filepath.SkipDir
				}
				if !d.IsDir() {
					return filepath.SkipDir
				}
				if !watchFilter.IsAccepted(absPath) {
					return filepath.SkipDir
				}
				if err := watcher.Add(absPath); err != nil {
					log.Printf("Error adding watch to %q: %v", absPath, err)
					return filepath.SkipDir
				}
				watchedPaths[absPath] = struct{}{}
				result[absPath] = struct{}{}
				return nil
			})
			if verbose {
				for path := range result {
					log.Printf("Watching %q", path)
				}
			}
			return result
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
				log.Printf("Filesystem watch event: %s", event)
			}
			path, err := filesystem.ToAbsolutePath(event.Name)
			if err != nil {
				log.Printf("Error processing path: %v", err)
				return
			}
			if !watchFilter.IsAccepted(path) {
				if verbose {
					log.Printf("Skipping excluded path %q from processing", path)
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
				log.Printf("Filesystem watcher error: %v", err)
			}
		}
	}
}
