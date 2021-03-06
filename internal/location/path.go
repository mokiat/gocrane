package location

import (
	"fmt"
	"path/filepath"
	"strings"
)

// MustParsePath is similar to ParsePath except that it panics
// if there is an error.
func MustParsePath(p string) Path {
	path, err := ParsePath(p)
	if err != nil {
		panic(fmt.Errorf("failed to parse path: %w", err))
	}
	return path
}

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

// String returns a string representation of this Path.
func (p Path) String() string {
	return fmt.Sprintf("%c%s", filepath.Separator, strings.Join(p, string(filepath.Separator)))
}
