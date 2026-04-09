package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/mokiat/gocrane/internal/filesystem"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/filter"
	"github.com/mokiat/gog/opt"
)

// NewWatchNode creates a new WatchNode with the specified arguments.
func NewWatchNode(rootDirs []string, watchFilter filter.Func[string], verbose bool) *WatchNode {
	return &WatchNode{
		rootDirs:    rootDirs,
		watchFilter: watchFilter,
		verbose:     verbose,

		trackedPaths: ds.NewSet[string](128),
	}
}

// WatchNode is responsible for watching the filesystem for changes and emitting
// change events accordingly.
type WatchNode struct {
	rootDirs    []string
	watchFilter filter.Func[string]
	verbose     bool

	watcher      *fsnotify.Watcher
	trackedPaths *ds.Set[string]
}

// Run starts the watch node.
func (n *WatchNode) Run(ctx context.Context, outEvents Queue[ChangeEvent]) error {
	var err error
	n.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create filesystem watcher: %w", err)
	}
	defer n.watcher.Close()

	for _, rootDir := range n.rootDirs {
		// don't output events when bootstrapping
		n.startWatching(ctx, rootDir, opt.Unspecified[Queue[ChangeEvent]]())
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-n.watcher.Events:
			if !n.handleFSWatchEvent(ctx, event, outEvents) {
				return nil
			}
		case err := <-n.watcher.Errors:
			n.logFSWatchError(err)
		}
	}
}

func (n *WatchNode) handleFSWatchEvent(ctx context.Context, event fsnotify.Event, outEvents Queue[ChangeEvent]) bool {
	n.logFSWatchEvent(event)

	absPath, err := filesystem.ToAbsolutePath(event.Name)
	if err != nil {
		n.logPathAbsConvertError(event.Name, err)
		return true // regardless, keep going
	}

	if !n.watchFilter(absPath) {
		n.logExcludedPathWatchSkip(absPath)
		return true // not relevant, ignore
	}

	switch {
	case event.Has(fsnotify.Create):
		return n.startWatching(ctx, absPath, opt.V(outEvents))

	case event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename):
		// Note: Rename is produced on Linux when a file is deleted.
		return n.stopWatching(ctx, absPath, opt.V(outEvents))

	case event.Has(fsnotify.Write):
		// Note: Always check Write before Chmod, since on MacOS both events
		// can be produced simultaneously for the same change.
		return outEvents.Push(ctx, ChangeEvent{Path: absPath})

	case event.Has(fsnotify.Chmod):
		// We do nothing on these, since MacOS produces a lot of them.
		return true

	default:
		// Don't trigger on unknown events.
		return true
	}
}

func (n *WatchNode) startWatching(ctx context.Context, root string, optOutEvents opt.T[Queue[ChangeEvent]]) bool {
	filesystem.Traverse(root, func(p string, isDir bool, err error) error {
		if err != nil {
			n.logTraverseError(p, err)
			return filesystem.ErrSkip
		}

		absPath, err := filesystem.ToAbsolutePath(p)
		if err != nil {
			n.logPathAbsConvertError(p, err)
			return filesystem.ErrSkip
		}

		if n.trackedPaths.Contains(absPath) {
			return filesystem.ErrSkip
		}

		if !n.watchFilter(absPath) {
			return filesystem.ErrSkip
		}

		if isDir {
			if err := n.watcher.Add(absPath); err != nil {
				n.logFSWatchAddError(absPath, err)
				return nil // continue traversal to still attempt subdirectories
			}
		}

		n.trackedPaths.Add(absPath)
		n.logStartWatching(absPath)

		if outEvents, ok := optOutEvents.Unwrap(); ok {
			if !outEvents.Push(ctx, ChangeEvent{Path: absPath}) {
				return ctx.Err()
			}
		}
		return nil
	})

	return ctx.Err() == nil
}

func (n *WatchNode) stopWatching(ctx context.Context, root string, optOutEvents opt.T[Queue[ChangeEvent]]) bool {
	// rootSubPath is used to ensure that we stop tracking only children of the root path,
	// and not sibling paths that share the same prefix
	rootSubPath := root + string(filepath.Separator)
	for p := range n.trackedPaths.Unbox() {
		if p == root || strings.HasPrefix(p, rootSubPath) {
			err := n.watcher.Remove(p)
			if err != nil && !errors.Is(err, fsnotify.ErrNonExistentWatch) {
				n.logFSWatchRemoveError(p, err)
				continue
			}

			n.trackedPaths.Remove(p)
			n.logStopWatching(p)

			if outEvents, ok := optOutEvents.Unwrap(); ok {
				if !outEvents.Push(ctx, ChangeEvent{Path: p}) {
					return false
				}
			}
		}
	}
	return true
}

func (n *WatchNode) logFSWatchEvent(event fsnotify.Event) {
	if n.verbose {
		log.Printf("Filesystem watch event: %s", event)
	}
}

func (n *WatchNode) logFSWatchError(err error) {
	log.Printf("Filesystem watch error: %v", err)
}

func (n *WatchNode) logFSWatchAddError(path string, err error) {
	log.Printf("Filesystem watch add(%q) error: %v", path, err)
}

func (n *WatchNode) logFSWatchRemoveError(path string, err error) {
	log.Printf("Filesystem watch remove(%q) error: %v", path, err)
}

func (n *WatchNode) logPathAbsConvertError(path string, err error) {
	log.Printf("Error converting path %q to absolute: %v", path, err)
}

func (n *WatchNode) logTraverseError(path string, err error) {
	log.Printf("Error traversing %q: %v", path, err)
}

func (n *WatchNode) logStartWatching(path string) {
	if n.verbose {
		log.Printf("Now watching %q", path)
	}
}

func (n *WatchNode) logStopWatching(path string) {
	if n.verbose {
		log.Printf("No longer watching %q", path)
	}
}

func (n *WatchNode) logExcludedPathWatchSkip(path string) {
	if n.verbose {
		log.Printf("Skipping excluded path %q from processing", path)
	}
}
