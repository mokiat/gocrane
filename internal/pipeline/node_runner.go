package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mokiat/gocrane/internal/project"
)

// RunRequest can be sent to the RunnerNode to trigger it to run a new binary.
type RunRequest struct {

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
func (r *RunnerNode) Run(ctx context.Context, requestQueue Queue[RunRequest]) error {
	var request RunRequest
	for requestQueue.Pop(ctx, &request) {
		if err := r.stopProcess(); err != nil {
			return err
		}
		if err := r.startProcess(request.BinaryPath); err != nil {
			return err
		}
	}
	return r.stopProcess()
}

func (r *RunnerNode) startProcess(binaryPath string) error {
	// Note: Intentionally not using the global Run context here, as we want to
	// manage the lifecycle of the process ourselves.
	if r.process != nil {
		return fmt.Errorf("there is already a running process") // should not happen
	}
	log.Printf("Starting new process...")
	process, err := r.runner.Run(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	r.process = process
	log.Printf("Successfully started new process.")
	return nil
}

func (r *RunnerNode) stopProcess() error {
	// Note: Intentionally not using the global Run context here, as we want to
	// manage the lifecycle of the process ourselves.
	if r.process == nil {
		return nil // will occur for the first build event, and is not an error
	}
	log.Printf("Stopping running process (timeout: %s)...", r.shutdownTimeout)
	ctxShutdown, forceShutdown := context.WithTimeout(context.Background(), r.shutdownTimeout)
	defer forceShutdown()
	if err := r.process.Stop(ctxShutdown); err != nil {
		return fmt.Errorf("failed to stop process: %w", err)
	}
	r.process = nil
	log.Printf("Successfully stopped running process.")
	return nil
}
