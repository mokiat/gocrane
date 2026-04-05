package pipeline

// CompileRequest can be sent to the CompilerNode to trigger it to compile a
// new binary.
type CompileRequest struct{}

// CompileResponse is emitted when a new binary has been successfully compiled.
type CompileResponse struct {

	// BinaryPath is the path to the newly built binary.
	BinaryPath string
}

// NewCompilerNode creates a new CompilerNode with the specified arguments.
func NewCompilerNode() *CompilerNode {
	return &CompilerNode{}
}

// CompilerNode is responsible for compiling binaries.
type CompilerNode struct {
	// TODO: Implement similar to RunnerNode.
}
