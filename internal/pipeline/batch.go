package pipeline

import (
	"context"
	"time"

	"github.com/mokiat/gocrane/internal/events"
)

func Batch(
	ctx context.Context,
	in events.ChangeQueue,
	out events.ChangeQueue,
	batchDuration time.Duration,
) func() error {

	return func() error {
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
				flushChan = out // Allow flushing.

			// Try to read new events and accumulate them.
			case event := <-in:
				// Start or extend flush timer on event received.
				timerChan = time.After(batchDuration)
				batchEvent.Paths = append(batchEvent.Paths, event.Paths...)
			}
		}
	}
}
