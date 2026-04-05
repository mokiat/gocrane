package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mokiat/gocrane/internal/project"
)

// BuilderInput can be sent to the BuilderNode to trigger it to build a
// new binary.
type BuilderInput struct{}

// BuilderOutput is emitted when a new binary has been successfully compiled.
type BuilderOutput struct {

	// BinaryPath is the path to the newly built binary.
	BinaryPath string
}

// NewBuilderNode creates a new BuilderNode with the specified arguments.
func NewBuilderNode(mainDir string, buildArgs []string) *BuilderNode {
	return &BuilderNode{
		builder: project.NewBuilder(mainDir, buildArgs),
	}
}

// BuilderNode is responsible for compiling binaries.
type BuilderNode struct {
	builder *project.Builder
}

// Run starts the builder node, which listens for new build events and compiles
// the corresponding binaries.
//
// If the context is canceled, the builder node will stop any ongoing
// compilation and exit.
func (n *BuilderNode) Run(ctx context.Context, inputs Queue[BuilderInput], outputs Queue[BuilderOutput]) error {
	tempDir, err := os.MkdirTemp("", "gocrane-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	var input BuilderInput
	for inputs.Pop(ctx, &input) {
		log.Printf("Building...")
		binaryName := fmt.Sprintf("executable-%s", uuid.NewString())
		binaryPath := filepath.Join(tempDir, binaryName)
		if err := n.builder.Build(ctx, binaryPath); err != nil {
			log.Printf("Build failure: %v", err)
			continue
		}
		log.Printf("Build was successful.")
		output := BuilderOutput{
			BinaryPath: binaryPath,
		}
		if !outputs.Push(ctx, output) {
			break
		}
	}
	return nil
}
