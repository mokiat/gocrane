package project

import (
	"fmt"
	"io"
	"os"
)

// OpenDigestFile reads the digest string from the specified file.
func OpenDigestFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", path, err)
	}
	return string(data), err
}

// SaveDigestFile stores the specified digest into the specified file.
func SaveDigestFile(path, digest string) error {
	if err := os.WriteFile(path, []byte(digest), 0x644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}
	return nil
}

// WriteFileDigest writes the digest of the specified file to the specified
// Writer.
func WriteFileDigest(out io.Writer, file string) error {
	stat, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to state file %q: %w", file, err)
	}
	// Note: Don't include millisecond precision, as that seems to differ between
	// host and client machine (in some cases it is not included).
	const timeFormat = "2006/01/02 15:04:05"
	fmt.Fprint(out, len(file), file, stat.ModTime().UTC().Format(timeFormat), stat.Size())
	return nil
}
