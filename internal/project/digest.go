package project

import (
	"fmt"
	"io"
	"os"
)

func ReadDigest(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read from file %q: %w", path, err)
	}
	return string(data), nil
}

func WriteDigest(path, digest string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(digest); err != nil {
		return fmt.Errorf("failed to write to file %q: %w", path, err)
	}
	return nil
}
