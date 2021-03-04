package location

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ParsePath attempts to parse the specified location into an absolute,
// cleaned Path.
func ParsePath(p string) (Path, error) {
	fp, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return nil, fmt.Errorf("failed to convert location to absolute: %w", err)
	}
	return strings.Split(fp, string(filepath.Separator)), nil
}

// Path consists of a slice of segments that make up an absolute path on the
// filesystem.
type Path []string
