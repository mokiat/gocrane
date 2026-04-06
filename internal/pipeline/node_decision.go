package pipeline

import "context"

// TODO: Write tests for this node.

// NewDecisionNode creates a new DecisionNode with the specified arguments.
func NewDecisionNode() *DecisionNode {
	// TODO: Take filter trees as arguments.
	return &DecisionNode{}
}

// DecisionNode is responsible for deciding whether a change event should
// trigger a restart event and whether that restart should include a rebuild
// as well.
type DecisionNode struct{}

// Run starts the decision node.
func (n *DecisionNode) Run(ctx context.Context, inEvents Queue[ChangeEvent], outEvents Queue[RestartEvent]) error {
	return nil
}
