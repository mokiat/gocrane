package project

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
)

func NewBuilder(runDir string) (*Builder, error) {
	tempDir, err := ioutil.TempDir("", "gocrane-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	return &Builder{
		runDir:  runDir,
		tempDir: tempDir,
	}, nil
}

type Builder struct {
	runDir  string
	tempDir string
}

func (b *Builder) Build(ctx context.Context) (string, error) {
	fileName := fmt.Sprintf("executable-%s", uuid.NewV4())
	path := filepath.Join(b.tempDir, fileName)

	cmd := exec.CommandContext(ctx, "go", "build", "-o", path, b.runDir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run go build: %w", err)
	}
	return path, nil
}

func (b *Builder) Cleanup() error {
	if err := os.RemoveAll(b.tempDir); err != nil {
		return fmt.Errorf("failed to delete temp directory: %w", err)
	}
	return nil
}
