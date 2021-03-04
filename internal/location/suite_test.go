package location_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLocation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Location Suite")
}

type Fixture struct {
	tempDir string
}

func (f *Fixture) Root() string {
	return f.tempDir
}

func (f *Fixture) Create() error {
	dir, err := os.MkdirTemp("", "gocrane-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	f.tempDir = dir
	return nil
}

func (f *Fixture) Delete() error {
	if f.tempDir == "" {
		return nil
	}
	if err := os.RemoveAll(f.tempDir); err != nil {
		return fmt.Errorf("failed to delete temp directory: %w", err)
	}
	return nil
}

func (f *Fixture) Path(path string) string {
	return filepath.Join(f.tempDir, path)
}

func (f *Fixture) CreateDir(path string) error {
	if err := os.MkdirAll(f.Path(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}
