package change

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/mokiat/gocrane/internal/events"
	"github.com/mokiat/gocrane/internal/project"
)

func NewWatcher(verbose bool, dirs project.FileSet, filter *project.Filter) *Watcher {
	return &Watcher{
		dirs:    dirs,
		filter:  filter,
		verbose: verbose,
	}
}

type Watcher struct {
	dirs    project.FileSet
	filter  *project.Filter
	verbose bool
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

func (w *Watcher) isExcludedPath(path string) bool {
	return w.filter.Match(path)
}

type watcherExecution struct {
	watcher     *Watcher
	fsWatcher   *fsnotify.Watcher
	changeQueue events.ChangeQueue
}

func (e *watcherExecution) Run(ctx context.Context) error {
	for path := range e.watcher.dirs {
		filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			path = filepath.Clean(path)
			if err != nil {
				e.recordError(fmt.Errorf("failed to traverse %q: %w", path, err))
				return filepath.SkipDir
			}
			if d.IsDir() {
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
	if e.watcher.isExcludedPath(path) {
		return
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
		if !e.watcher.isExcludedPath(path) {
			e.recordEvent(events.Change{
				Paths: []string{path},
			})
		}
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
