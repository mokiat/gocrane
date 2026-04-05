package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mokiat/gocrane/internal/project"
)

// RunnerInput can be sent to the RunnerNode to trigger it to run a new binary.
type RunnerInput struct {

	// BinaryPath is the path to the newly built binary that should be run.
	BinaryPath string
}

// NewRunnerNode creates a new RunnerNode with the specified arguments.
func NewRunnerNode(workDir string, runArgs []string, shutdownTimeout time.Duration) *RunnerNode {
	return &RunnerNode{
		runner:          project.NewRunner(workDir, runArgs),
		process:         nil,
		shutdownTimeout: shutdownTimeout,
	}
}

// RunnerNode is responsible for running a binary and stopping it when a new
// one is available or when the context is canceled.
type RunnerNode struct {
	runner          *project.Runner
	process         *project.Process
	shutdownTimeout time.Duration
}

// Run starts the runner node, which listens for new build events and runs the
// corresponding binaries.
//
// If the context is canceled, the runner node will stop any running process
// and exit.
func (n *RunnerNode) Run(ctx context.Context, inputs Queue[RunnerInput]) error {
	var input RunnerInput
	for inputs.Pop(ctx, &input) {
		if err := n.stopProcess(); err != nil {
			return err
		}
		if err := n.startProcess(input.BinaryPath); err != nil {
			return err
		}
	}
	return n.stopProcess()
}

func (n *RunnerNode) startProcess(binaryPath string) error {
	// Note: Intentionally not using the global Run context here, as we want to
	// manage the lifecycle of the process ourselves.
	if n.process != nil {
		return fmt.Errorf("there is already a running process") // should not happen
	}
	log.Printf("Starting new process...")
	process, err := n.runner.Run(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	n.process = process
	log.Printf("Successfully started new process.")
	return nil
}

func (n *RunnerNode) stopProcess() error {
	// Note: Intentionally not using the global Run context here, as we want to
	// manage the lifecycle of the process ourselves.
	if n.process == nil {
		return nil // will occur for the first build event, and is not an error
	}
	log.Printf("Stopping running process (timeout: %s)...", n.shutdownTimeout)
	ctxShutdown, forceShutdown := context.WithTimeout(context.Background(), n.shutdownTimeout)
	defer forceShutdown()
	if err := n.process.Stop(ctxShutdown); err != nil {
		return fmt.Errorf("failed to stop process: %w", err)
	}
	n.process = nil
	log.Printf("Successfully stopped running process.")
	return nil
}
