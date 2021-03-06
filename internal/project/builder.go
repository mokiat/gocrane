package project

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/mokiat/gocrane/internal/logutil"
)

func NewBuilder(runDir string, args []string, destination string) *Builder {
	return &Builder{
		runDir:      runDir,
		args:        args,
		destination: destination,
	}
}

type Builder struct {
	runDir      string
	args        []string
	destination string
}

func (b *Builder) Build(ctx context.Context) error {
	absDestination, err := filepath.Abs(b.destination)
	if err != nil {
		return fmt.Errorf("failed to get absolute destination for %q: %w", b.destination, err)
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
