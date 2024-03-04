package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/mokiat/gocrane/internal/filesystem"
	"github.com/mokiat/gog/ds"
)

func Watch(
	ctx context.Context,
	verbose bool,
	dirs []string,
	watchFilter *filesystem.FilterTree,
	out Queue[ChangeEvent],
	bootstrapEvent *ChangeEvent,

) func() error {

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

		proc := &watchProcess{
			verbose:      verbose,
			watcher:      watcher,
			watchFilter:  watchFilter,
			trackedPaths: ds.NewSet[string](1024),
		}

		// Bootstrap watching.
		for _, dir := range dirs {
			proc.startWatching(dir)
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case event := <-watcher.Events:
				changedPaths := proc.handleEvent(event)
				if changedPaths != nil && !changedPaths.IsEmpty() {
					out.Push(ctx, ChangeEvent{
						Paths: changedPaths.Items(),
					})
				}
			case err := <-watcher.Errors:
				proc.logFSWatchError(err)
			}
		}
	}
}

type watchProcess struct {
	verbose     bool
	watcher     *fsnotify.Watcher
	watchFilter *filesystem.FilterTree

	trackedPaths *ds.Set[string]
}

func (proc *watchProcess) handleEvent(event fsnotify.Event) *ds.Set[string] {
	proc.logFSWatchEvent(event)

	absPath, err := filesystem.ToAbsolutePath(event.Name)
	if err != nil {
		proc.logPathAbsConvertError(event.Name, err)
		return nil
	}

	if !proc.shouldTrack(absPath) {
		proc.logExcludedPathWatchSkip(absPath)
		return nil
	}

	switch {
	case event.Has(fsnotify.Create):
		return proc.startWatching(absPath)

	case event.Has(fsnotify.Rename):
		// Rename is produced on Linux when a file is deleted.
		return proc.stopWatching(absPath)

	case event.Has(fsnotify.Remove):
		return proc.stopWatching(absPath)

	case event.Has(fsnotify.Chmod):
		// We do nothing on these, since MacOS produces a lot of them.
		return nil

	default:
		return ds.SetFromSlice([]string{absPath})
	}
}

func (proc *watchProcess) startWatching(root string) *ds.Set[string] {
	result := ds.NewSet[string](1)

	filesystem.Traverse(root, func(p string, isDir bool, err error) error {
		if err != nil {
			proc.logTraverseError(p, err)
			return filesystem.ErrSkip
		}

		absPath, err := filesystem.ToAbsolutePath(p)
		if err != nil {
			proc.logPathAbsConvertError(p, err)
			return filesystem.ErrSkip
		}

		if proc.isTracked(absPath) {
			return filesystem.ErrSkip
		}

		if !proc.shouldTrack(absPath) {
			return filesystem.ErrSkip
		}

		if isDir {
			if err := proc.watcher.Add(absPath); err != nil {
				proc.logFSWatchAddError(absPath, err)
				return filesystem.ErrSkip
			}
		}

		proc.trackPath(absPath)
		result.Add(absPath)
		return nil
	})

	for path := range result.Unbox() {
		proc.logStartWatching(path)
	}
	return result
}

func (proc *watchProcess) stopWatching(root string) *ds.Set[string] {
	result := ds.NewSet[string](1)

	for p := range proc.trackedPaths.Unbox() {
		if strings.HasPrefix(p, root) {
			result.Add(p)
			err := proc.watcher.Remove(p)
			if err == nil || errors.Is(err, fsnotify.ErrNonExistentWatch) {
				proc.untrackPath(p)
			} else {
				proc.logFSWatchRemoveError(p, err)
			}
		}
	}

	for path := range result.Unbox() {
		proc.logStopWatching(path)
	}
	return result
}

func (proc *watchProcess) shouldTrack(path string) bool {
	return proc.watchFilter.IsAccepted(path)
}

func (proc *watchProcess) trackPath(path string) {
	proc.trackedPaths.Add(path)
}

func (proc *watchProcess) untrackPath(path string) {
	proc.trackedPaths.Remove(path)
}

func (proc *watchProcess) isTracked(path string) bool {
	return proc.trackedPaths.Contains(path)
}

func (proc *watchProcess) logFSWatchEvent(event fsnotify.Event) {
	if proc.verbose {
		log.Printf("Filesystem watch event: %s", event)
	}
}

func (proc *watchProcess) logFSWatchAddError(path string, err error) {
	log.Printf("Error adding watch to %q: %v", path, err)
}

func (proc *watchProcess) logFSWatchRemoveError(path string, err error) {
	log.Printf("Error removing watch from %q: %v", path, err)
}

func (proc *watchProcess) logFSWatchError(err error) {
	log.Printf("Filesystem watch error: %v", err)
}

func (proc *watchProcess) logStartWatching(path string) {
	if proc.verbose {
		log.Printf("Now watching %q", path)
	}
}

func (proc *watchProcess) logStopWatching(path string) {
	if proc.verbose {
		log.Printf("No longer watching %q", path)
	}
}

func (proc *watchProcess) logTraverseError(path string, err error) {
	log.Printf("Error traversing %q: %v", path, err)
}

func (proc *watchProcess) logPathAbsConvertError(path string, err error) {
	log.Printf("Error converting path %q to absolute: %v", path, err)
}

func (proc *watchProcess) logExcludedPathWatchSkip(path string) {
	if proc.verbose {
		log.Printf("Skipping excluded path %q from processing", path)
	}
}
