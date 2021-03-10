package pipeline

import "context"

type ChangeEvent struct {
	Paths []string
}

type ChangeEventQueue chan ChangeEvent

func (q ChangeEventQueue) Push(ctx context.Context, event ChangeEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case q <- event:
		return true
	}
}

func (q ChangeEventQueue) Pop(ctx context.Context, ptr *ChangeEvent) bool {
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

type BuildEvent struct {
	Path string
}

type BuildEventQueue chan BuildEvent

func (q BuildEventQueue) Push(ctx context.Context, event BuildEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case q <- event:
		return true
	}
}

func (q BuildEventQueue) Pop(ctx context.Context, ptr *BuildEvent) bool {
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
