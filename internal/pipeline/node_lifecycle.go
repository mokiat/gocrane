package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mokiat/gocrane/internal/project"
	"github.com/mokiat/gog/opt"
)

// NewLifecycleNode creates a new LifecycleNode with the specified arguments.
func NewLifecycleNode(builder *project.Builder, runner *project.Runner, cachedBinary opt.T[string], shutdownTimeout time.Duration) *LifecycleNode {
	return &LifecycleNode{
		builder:         builder,
		runner:          runner,
		cachedBinary:    cachedBinary,
		shutdownTimeout: shutdownTimeout,
		process:         nil,
	}
}

// LifecycleNode is responsible for orchestrating the builder and runner nodes.
type LifecycleNode struct {
	builder         *project.Builder
	runner          *project.Runner
	cachedBinary    opt.T[string]
	shutdownTimeout time.Duration
	process         *project.Process
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
func (n *LifecycleNode) Run(ctx context.Context, events Queue[RestartEvent]) error {
	tempDir, err := os.MkdirTemp("", "gocrane-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	currentBinaryPath := n.cachedBinary.ValueOrDefault("")

	var event RestartEvent
	for events.Pop(ctx, &event) {
		if event.ShouldRebuild || (currentBinaryPath == "") {
			binaryPath, err := n.buildBinary(ctx, tempDir)
			if err != nil {
				continue // keep the old binary running
			}
			currentBinaryPath = binaryPath
		}
		if err := n.stopProcess(); err != nil {
			return err
		}
		if ctx.Err() != nil {
			break
		}
		if err := n.startProcess(currentBinaryPath); err != nil {
			return err
		}
	}
	return n.stopProcess()
}

func (n *LifecycleNode) buildBinary(ctx context.Context, tempDir string) (string, error) {
	log.Printf("Building...")
	binaryName := fmt.Sprintf("executable-%s", uuid.NewString())
	binaryPath := filepath.Join(tempDir, binaryName)
	if err := n.builder.Build(ctx, binaryPath); err != nil {
		log.Printf("Build failure: %v", err)
		return "", err
	}
	log.Printf("Build was successful.")
	return binaryPath, nil
}

func (n *LifecycleNode) startProcess(binaryPath string) error {
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

func (n *LifecycleNode) stopProcess() error {
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
