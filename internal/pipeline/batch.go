package pipeline

import (
	"context"
	"time"
)

func Batch(
	ctx context.Context,
	in Queue[ChangeEvent],
	out Queue[ChangeEvent],
	batchDuration time.Duration,
) func() error {

	return func() error {
		var (
			flushTimer                    = time.NewTimer(batchDuration)
			flushChan  chan<- ChangeEvent = nil
			batchEvent ChangeEvent
		)

		stopTimer := func() {
			if !flushTimer.Stop() {
				select {
				case <-flushTimer.C:
				default:
				}
			}
		}

		for {
			select {
			// Check if we should exit.
			case <-ctx.Done():
				stopTimer()
				return nil

			// Check if we are able to flush. If we cannot push the batched event,
			// either because flushing is disabled or the receiver is blocked, we will
			// continue to accumulate batched events.
			case flushChan <- batchEvent:
				flushChan = nil        // Disable flushing.
				batchEvent.Paths = nil // Don't reuse the slice!

			// A sufficient amount of time has passed since the first event was received
			// so we can enable flushing.
			case <-flushTimer.C:
				flushChan = out // Allow flushing.

			// Try to read new events and accumulate them.
			case event := <-in:
				batchEvent.Paths = append(batchEvent.Paths, event.Paths...)
				stopTimer()
				flushTimer.Reset(batchDuration)
			}
		}
	}
}
