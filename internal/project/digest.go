package project

import (
	"fmt"
	"os"
)

// ReadDigest reads the digest string from the specified file.
func ReadDigest(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", path, err)
	}
	return string(data), err
}

// WriteDigest stores the specified digest into the specified file.
func WriteDigest(path, digest string) error {
	if err := os.WriteFile(path, []byte(digest), 0x644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}
	return nil
}
