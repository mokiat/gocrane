package events

import "context"

type Build struct {
	Path string
}

type BuildQueue chan Build

func (q BuildQueue) Pop(ctx context.Context, ptr *Build) bool {
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

func (q BuildQueue) Push(ctx context.Context, event Build) {
	select {
	case <-ctx.Done():
	case q <- event:
	}
}
