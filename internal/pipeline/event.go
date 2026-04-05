package pipeline

import "context"

type ChangeEvent struct {
	Paths []string
}

// RunEvent is emitted when a new binary is ready to be run.
type RunEvent struct {

	// BinaryPath is the path to the newly built binary that should be run.
	BinaryPath string
}

// Queue is a simple wrapper around a channel that provides push and pop methods
// with context support.
type Queue[T any] chan T

// Push attempts to push a value into the queue, and returns false if the
// context is canceled before the value is pushed.
func (q Queue[T]) Push(ctx context.Context, value T) bool {
	select {
	case <-ctx.Done():
		return false
	case q <- value:
		return true
	}
}

// Pop attempts to pop a value from the queue, and returns false if the context
// is canceled before a value is popped.
func (q Queue[T]) Pop(ctx context.Context, ptr *T) bool {
	select {
	case event, ok := <-q:
		if !ok {
			return false
		}
		*ptr = event
		return true
	case <-ctx.Done():
		return false
	}
}
