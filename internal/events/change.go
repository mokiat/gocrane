package events

import (
	"context"
)

type Change struct {
	Paths []string
}

type ChangeQueue chan Change

func (q ChangeQueue) Push(ctx context.Context, event Change) {
	select {
	case <-ctx.Done():
	case q <- event:
	}
}

func (q ChangeQueue) Pop(ctx context.Context, ptr *Change) bool {
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

func (q ChangeQueue) Watch(ctx context.Context, fn func(event Change) error) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-q:
			if !ok {
				return nil
			}
			if err := fn(event); err != nil {
				return err
			}
		}
	}
}
