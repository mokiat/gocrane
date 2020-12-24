package change

import (
	"context"
	"time"

	"github.com/mokiat/gocrane/internal/events"
)

func NewBatcher(inactivity time.Duration) *Batcher {
	return &Batcher{
		inactivity: inactivity,
	}
}

type Batcher struct {
	inactivity time.Duration
}

func (b *Batcher) Run(ctx context.Context, changeEventQueue, batchChangeEventQueue events.ChangeQueue) error {
	var (
		flushChan  chan<- events.Change = nil
		timerChan  <-chan time.Time     = nil
		batchEvent events.Change
	)

	for {
		select {
		// Check if we should exit.
		case <-ctx.Done():
			return nil

		// Check if we are able to flush. If we cannot push the batched event,
		// either because flushing is disabled or the receiver is blocked, we will
		// continue to accumulate batched events.
		case flushChan <- batchEvent:
			flushChan = nil        // Disable flushing.
			batchEvent.Paths = nil // Don't reuse the slice!

		// A sufficient amount of time has passed since the first event was received
		// so we can enable flushing.
		case <-timerChan:
			timerChan = nil
			flushChan = batchChangeEventQueue // Allow flushing.

		// Try to read new events and accumulate them.
		case event := <-changeEventQueue:
			// Start or extend flush timer on event received.
			timerChan = time.After(b.inactivity) // TODO: Use timer instead!
			batchEvent.Paths = append(batchEvent.Paths, event.Paths...)
		}
	}
}
