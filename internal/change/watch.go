package change

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/mokiat/gocrane/internal/events"
	"github.com/mokiat/gocrane/internal/location"
)

func NewWatcher(verbose bool, dirs []string, filter location.Filter) *Watcher {
	return &Watcher{
		dirs:    dirs,
		filter:  filter,
		verbose: verbose,
	}
}

type Watcher struct {
	dirs    []string
	filter  location.Filter
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
		filter:      w.filter,
	}
	return execution.Run(ctx)
}

type watcherExecution struct {
	watcher     *Watcher
	fsWatcher   *fsnotify.Watcher
	changeQueue events.ChangeQueue
	filter      location.Filter
}

func (e *watcherExecution) Run(ctx context.Context) error {
	for _, dir := range e.watcher.dirs {
		e.considerDir(dir)
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

func (e *watcherExecution) considerDir(path string) {
	if !e.filter.Match(path) {
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
	path, err := filepath.Abs(event.Name)
	if err != nil {
		e.recordError(fmt.Errorf("failed to conver path to absolute %q: %w", event.Name, err))
		return
	}
	if !e.filter.Match(path) {
		return
	}
	if isEventType(event, fsnotify.Create) {
		e.considerDir(path)
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
