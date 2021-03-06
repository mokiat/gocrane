package project

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/mokiat/gocrane/internal/logutil"
)

func NewBuilder(runDir string, args []string) *Builder {
	return &Builder{
		runDir: runDir,
		args:   args,
	}
}

type Builder struct {
	runDir string
	args   []string
}

func (b *Builder) Build(ctx context.Context, destination string) error {
	absDestination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("failed to get absolute destination for %q: %w", destination, err)
	}

	args := append([]string{"build"}, b.args...)
	args = append(args, "-o", absDestination, "./")

	logger := log.New(log.Writer(), "[compiler]: ", log.Ltime|log.Lmsgprefix)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = b.runDir
	cmd.Stdout = logutil.ToWriter(logger)
	cmd.Stderr = logutil.ToWriter(logger)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go build: %w", err)
	}
	return nil
}
