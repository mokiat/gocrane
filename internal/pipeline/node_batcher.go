package pipeline

import (
	"context"
	"time"
)

// NewBatcherNode creates a new BatcherNode with the specified batch duration.
func NewBatcherNode(batchDuration time.Duration) *BatcherNode {
	return &BatcherNode{
		batchDuration: batchDuration,
	}
}

// BatcherNode is responsible for batching incoming events and emitting a single
// restart event after a certain period of inactivity. This is useful to prevent
// excessive restarts when multiple file changes occur in quick succession.
type BatcherNode struct {
	batchDuration time.Duration

	flushTimer *time.Timer
	flushChan  chan<- RestartEvent
	batchEvent RestartEvent
}

// Run starts the batcher node.
func (n *BatcherNode) Run(ctx context.Context, inEvents, outEvents Queue[RestartEvent]) error {
	n.createTimer()
	defer n.stopTimer()

	for {
		select {
		case <-ctx.Done():
			return nil

		case event := <-inEvents:
			n.accumulateEvent(event)
			n.resetTimer()

		case <-n.flushTimer.C:
			n.allowFlush(outEvents)

		case n.flushChan <- n.batchEvent:
			n.clearEvent()
			n.disableFlush()
			n.stopTimer()
		}
	}
}

func (n *BatcherNode) createTimer() {
	n.flushTimer = time.NewTimer(n.batchDuration)
	n.flushTimer.Stop() // don't start it yet
}

func (n *BatcherNode) stopTimer() {
	n.flushTimer.Stop()
}

func (n *BatcherNode) resetTimer() {
	n.flushTimer.Reset(n.batchDuration)
}

func (n *BatcherNode) allowFlush(outEvents Queue[RestartEvent]) {
	n.flushChan = outEvents
}

func (n *BatcherNode) disableFlush() {
	// Note: A nil chan is never ready for sending, so this effectively disables
	// flushing until the next time we allow it.
	n.flushChan = nil
}

func (n *BatcherNode) clearEvent() {
	n.batchEvent = RestartEvent{}
}

func (n *BatcherNode) accumulateEvent(event RestartEvent) {
	n.batchEvent.ShouldRebuild = n.batchEvent.ShouldRebuild || event.ShouldRebuild
}
