package pipeline

import (
	"context"

	"github.com/mokiat/gog/filter"
	"github.com/mokiat/gog/opt"
)

// NewDecisionNode creates a new DecisionNode with the specified arguments.
func NewDecisionNode(rebuildFilter, restartFilter filter.Func[string]) *DecisionNode {
	return &DecisionNode{
		rebuildFilter: rebuildFilter,
		restartFilter: restartFilter,
	}
}

// DecisionNode is responsible for deciding whether a change event should
// trigger a restart event and whether that restart should include a rebuild
// as well.
type DecisionNode struct {
	rebuildFilter filter.Func[string]
	restartFilter filter.Func[string]
}

// Run starts the decision node.
func (n *DecisionNode) Run(ctx context.Context, inEvents Queue[ChangeEvent], outEvents Queue[RestartEvent]) error {
	var changeEvent ChangeEvent
	for inEvents.Pop(ctx, &changeEvent) {
		var optOutEvent opt.T[RestartEvent]
		switch {
		case n.rebuildFilter(changeEvent.Path):
			optOutEvent = opt.V(RestartEvent{
				ShouldRebuild: true,
			})
		case n.restartFilter(changeEvent.Path):
			optOutEvent = opt.V(RestartEvent{
				ShouldRebuild: false,
			})
		}
		if outEvent, ok := optOutEvent.Unwrap(); ok {
			if !outEvents.Push(ctx, outEvent) {
				return nil
			}
		}
	}
	return nil
}
