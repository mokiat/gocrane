package pipeline

import (
	"context"
	"time"

	"github.com/mokiat/gocrane/internal/change"
	"github.com/mokiat/gocrane/internal/events"
)

func Batch(
	ctx context.Context,
	in events.ChangeQueue,
	out events.ChangeQueue,
	batchDuration time.Duration,
) func() error {

	batcher := change.NewBatcher(batchDuration)

	return func() error {
		return batcher.Run(ctx, in, out)

	}
}
