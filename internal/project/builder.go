package project

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/mokiat/gocrane/internal/logutil"
)

// NewBuilder creates a new Builder with the specified main directory and build arguments.
func NewBuilder(mainDir string, args []string) *Builder {
	return &Builder{
		mainDir: mainDir,
		args:    args,
	}
}

// Builder is responsible for building the application binary.
type Builder struct {
	mainDir string
	args    []string
}

// Build builds the application binary and saves it to the specified destination file.
func (b *Builder) Build(ctx context.Context, outputFile string) error {
	absOutputFile, err := filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %q: %w", outputFile, err)
	}

	args := append([]string{"build"}, b.args...)
	args = append(args, "-o", absOutputFile, ".")

	logger := log.New(log.Writer(), "[compiler]: ", log.Ltime|log.Lmsgprefix)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = b.mainDir
	cmd.Stdout = logutil.ToWriter(logger)
	cmd.Stderr = logutil.ToWriter(logger)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to perform go build: %w", err)
	}
	return nil
}
