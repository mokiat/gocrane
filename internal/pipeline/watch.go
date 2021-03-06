package pipeline

import (
	"context"

	"github.com/mokiat/gocrane/internal/change"
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
	watcher := change.NewWatcher(verbose, dirs, watchFilter)

	return func() error {
		if bootstrapEvent != nil {
			out.Push(ctx, *bootstrapEvent)
		}
		return watcher.Run(ctx, out)
	}
}
