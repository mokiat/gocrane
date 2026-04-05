package pipeline

import (
	"context"

	"github.com/mokiat/gog/opt"
	"golang.org/x/sync/errgroup"
)

// LifecycleInput is sent to the LifecycleNode to trigger a refresh of the
// potentially running application.
type LifecycleInput struct {
	ShouldRebuild bool
}

// NewLifecycleNode creates a new LifecycleNode with the specified arguments.
func NewLifecycleNode(builderNode *BuilderNode, runnerNode *RunnerNode, cachedBinary opt.T[string]) *LifecycleNode {
	return &LifecycleNode{
		builderNode:  builderNode,
		runnerNode:   runnerNode,
		cachedBinary: cachedBinary,
	}
}

// LifecycleNode is responsible for orchestrating the builder and runner nodes.
type LifecycleNode struct {
	builderNode  *BuilderNode
	runnerNode   *RunnerNode
	cachedBinary opt.T[string]
}

// Run starts the lifecycle node, which orchestrates the builder and runner nodes.
//
// The lifecycle node listens for new lifecycle inputs. When a new input is
// received, it checks if a rebuild is requested or if there isn't a binary
// available yet. If either of those conditions is true, it triggers the builder
// node to build a new binary. Otherwise, it triggers the runner node to run the
// current binary.
//
// If the context is canceled, the lifecycle node will stop all ongoing work
// and exit.
func (n *LifecycleNode) Run(ctx context.Context, inputs Queue[LifecycleInput]) error {
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
		for {
			select {
			// If the context is canceled, exit this goroutine.
			case <-ctxGroup.Done():
				return nil

			// If a new build output is available, update the current binary and
			// trigger a run.
			case buildOutput := <-buildOutputs:
				currentBinary = buildOutput.BinaryPath
				runInput := RunnerInput{
					BinaryPath: currentBinary,
				}
				if !runInputs.Push(ctxGroup, runInput) {
					return nil
				}

			case input := <-inputs:
				// If a new lifecycle input is available, check if a rebuild is
				// requested or if there isn't a binary yet. If either is the case,
				// trigger a build. Otherwise, trigger a run with the current binary.
				if input.ShouldRebuild || (currentBinary == "") {
					buildInput := BuilderInput{}
					if !buildInputs.Push(ctxGroup, buildInput) {
						return nil
					}
				} else {
					runInput := RunnerInput{
						BinaryPath: currentBinary,
					}
					if !runInputs.Push(ctxGroup, runInput) {
						return nil
					}
				}
			}
		}
	})

	return group.Wait()
}
