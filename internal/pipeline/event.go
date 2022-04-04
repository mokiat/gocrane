package pipeline

import "context"

type Queue[T any] chan T

func (q Queue[T]) Push(ctx context.Context, value T) bool {
	select {
	case <-ctx.Done():
		return false
	case q <- value:
		return true
	}
}

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

type ChangeEvent struct {
	Paths []string
}

type BuildEvent struct {
	Path string
}
