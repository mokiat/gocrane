package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log"
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

			filesystem.Traverse(root, func(p string, isDir bool, err error) error {
				if err != nil {
					log.Printf("Error traversing %q: %v", p, err)
					return filesystem.ErrSkip
				}
				absPath, err := filesystem.ToAbsolutePath(p)
				if err != nil {
					log.Printf("Error converting path %q to absolute: %v", p, err)
					return filesystem.ErrSkip
				}
				if _, ok := watchedPaths[absPath]; ok {
					return filesystem.ErrSkip
				}
				if !watchFilter.IsAccepted(absPath) {
					return filesystem.ErrSkip
				}
				if isDir {
					if err := watcher.Add(absPath); err != nil {
						log.Printf("Error adding watch to %q: %v", absPath, err)
						return filesystem.ErrSkip
					}
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
					// NOTE: Regardless what the documentation says, we NEED to
					// try and explicitly Remove the watch as otherwise there are some
					// race condition bugs.
					//
					// Example on Linux:
					// 1. Create folder ./foo
					// 2. Create file ./foo/bar
					// 3. Delete folder ./foo
					// 4. Create folder ./foo
					// 5. Create file ./foo/bar
					// The file `bar` is indicated as being located at `/bar`
					// by the watcher. This bug does not appear of we remove the
					// watch explicitly.
					err := watcher.Remove(p)
					if err == nil || errors.Is(err, fsnotify.ErrNonExistentWatch) {
						delete(watchedPaths, p)
						if verbose {
							log.Printf("Unwatched %q", p)
						}
					} else {
						log.Printf("Error removing watch from %q: %v", p, err)
					}
				}
			}
			return result
		}

		processFSEvent := func(event fsnotify.Event) {
			if verbose {
				log.Printf("Filesystem watch event: %s", event)
			}
			absPath, err := filesystem.ToAbsolutePath(event.Name)
			if err != nil {
				log.Printf("Error processing path: %v", err)
				return
			}
			if !watchFilter.IsAccepted(absPath) {
				if verbose {
					log.Printf("Skipping excluded path %q from processing", absPath)
				}
				return
			}

			switch {
			case isEventType(event, fsnotify.Create):
				paths := watchPath(absPath)
				event := ChangeEvent{
					Paths: make([]string, 0, len(paths)),
				}
				for path := range paths {
					event.Paths = append(event.Paths, path)
				}
				out.Push(ctx, event)

			case isEventType(event, fsnotify.Rename):
				// Rename is produced on Linux when a file is deleted.
				paths := unwatchPath(absPath)
				event := ChangeEvent{
					Paths: make([]string, 0, len(paths)),
				}
				for path := range paths {
					event.Paths = append(event.Paths, path)
				}
				out.Push(ctx, event)

			case isEventType(event, fsnotify.Remove):
				paths := unwatchPath(absPath)
				event := ChangeEvent{
					Paths: make([]string, 0, len(paths)),
				}
				for path := range paths {
					event.Paths = append(event.Paths, path)
				}
				out.Push(ctx, event)

			case isEventType(event, fsnotify.Chmod):
				// We do nothing on these, since MacOS produces a lot of them.

			default:
				out.Push(ctx, ChangeEvent{
					Paths: []string{absPath},
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
