package change

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/mokiat/gocrane/internal/events"
)

func NewWatcher(includePaths, excludePaths []string, verbose bool) *Watcher {
	includePathSet := make(map[string]struct{})
	for _, path := range includePaths {
		includePathSet[filepath.Clean(path)] = struct{}{}
	}

	excludePathSet := make(map[string]struct{})
	for _, path := range excludePaths {
		excludePathSet[filepath.Clean(path)] = struct{}{}
	}

	return &Watcher{
		includePaths: includePathSet,
		excludePaths: excludePathSet,
		verbose:      verbose,
	}
}

type Watcher struct {
	includePaths map[string]struct{}
	excludePaths map[string]struct{}
	verbose      bool
}

func (w *Watcher) Run(ctx context.Context, changeEventQueue events.ChangeQueue) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create filesystem watcher: %w", err)
	}
	defer watcher.Close()

	execution := &watcherExecution{
		watcher:     w,
		fsWatcher:   watcher,
		changeQueue: changeEventQueue,
	}
	return execution.Run(ctx)
}

type watcherExecution struct {
	watcher     *Watcher
	changeQueue events.ChangeQueue
	fsWatcher   *fsnotify.Watcher
}

func (e *watcherExecution) Run(ctx context.Context) error {
	for path := range e.watcher.includePaths {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			path = filepath.Clean(path)
			if err != nil {
				e.recordError(fmt.Errorf("failed to traverse %q: %w", path, err))
				return filepath.SkipDir
			}
			if info.IsDir() {
				e.considerDir(path)
			}
			return nil
		})
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-e.fsWatcher.Events:
			e.processFSEvent(event)
		case err := <-e.fsWatcher.Errors:
			e.recordError(fmt.Errorf("filesystem watcher error: %w", err))
		}
	}
}

func (e *watcherExecution) considerPath(path string) {
	info, err := os.Stat(path)
	if err != nil {
		e.recordError(fmt.Errorf("failed to stat %q: %w", path, err))
		return
	}
	if info.IsDir() {
		e.considerDir(path)
	}
}

func (e *watcherExecution) considerDir(path string) {
	for excludedPath := range e.watcher.excludePaths {
		if strings.HasPrefix(path, excludedPath) {
			return
		}
	}

	if err := e.fsWatcher.Add(path); err != nil {
		e.recordError(fmt.Errorf("failed to watch %q: %w", path, err))
	} else {
		if e.watcher.verbose {
			log.Printf("watching: %q\n", path)
		}
	}
}

func (e *watcherExecution) processFSEvent(event fsnotify.Event) {
	if e.watcher.verbose {
		log.Printf("filesystem watch event: %s\n", event)
	}
	path := filepath.Clean(event.Name)
	if isEventType(event, fsnotify.Create) {
		e.considerPath(path)
	}
	if !isEventType(event, fsnotify.Chmod) {
		e.recordEvent(events.Change{
			Paths: []string{path},
		})
	}
}

func (e *watcherExecution) recordEvent(event events.Change) {
	select {
	case e.changeQueue <- event:
	default:
		log.Println("warning: event buffer full")
	}
}

func (e *watcherExecution) recordError(err error) {
	log.Printf("watcher error: %s", err)
}

func isEventType(event fsnotify.Event, eType fsnotify.Op) bool {
	return event.Op&eType == eType
}
