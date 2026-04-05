package pipeline

import (
	"context"

	"github.com/mokiat/gog/opt"
	"golang.org/x/sync/errgroup"
)

// NewInvalidationNode creates a new InvalidationNode with the specified arguments.
func NewInvalidationNode(builderNode *BuilderNode, runnerNode *RunnerNode, cachedBinary opt.T[string]) *InvalidationNode {
	return &InvalidationNode{
		builderNode:  builderNode,
		runnerNode:   runnerNode,
		cachedBinary: cachedBinary,
	}
}

// InvalidationNode is responsible for orchestrating the builder and runner nodes.
type InvalidationNode struct {
	builderNode  *BuilderNode
	runnerNode   *RunnerNode
	cachedBinary opt.T[string]
}

// Run starts the invalidation node, which orchestrates the builder and runner nodes.
//
// The invalidation node listens for new invalidation inputs. When a new input is
// received, it checks if a rebuild is requested or if there isn't a binary
// available yet. If either of those conditions is true, it triggers the builder
// node to build a new binary. Otherwise, it triggers the runner node to run the
// current binary.
//
// If the context is canceled, the invalidation node will stop all ongoing work
// and exit.
func (n *InvalidationNode) Run(ctx context.Context, events Queue[InvalidationEvent]) error {
	buildInputs := make(Queue[BuilderInput])
	buildOutputs := make(Queue[BuilderOutput])
	runInputs := make(Queue[RunnerInput])

	group, ctxGroup := errgroup.WithContext(ctx)

	// Start the builder node.
	group.Go(func() error {
		return n.builderNode.Run(ctxGroup, buildInputs, buildOutputs)
	})

	// Start the runner node.
	group.Go(func() error {
		return n.runnerNode.Run(ctxGroup, runInputs)
	})

	// Run main orchestration logic.
	group.Go(func() error {
		currentBinary := n.cachedBinary.ValueOrDefault("")

		var event InvalidationEvent
		for events.Pop(ctxGroup, &event) {
			if event.ShouldRebuild || (currentBinary == "") {
				buildInput := BuilderInput{}
				if !buildInputs.Push(ctxGroup, buildInput) {
					return nil
				}
				var buildOutput BuilderOutput
				if !buildOutputs.Pop(ctxGroup, &buildOutput) {
					return nil
				}
				currentBinary = buildOutput.BinaryPath
			}
			runInput := RunnerInput{
				BinaryPath: currentBinary,
			}
			if !runInputs.Push(ctxGroup, runInput) {
				return nil
			}
		}

		return nil
	})

	return group.Wait()
}
