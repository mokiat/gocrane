package filesystem

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ParsePath constructs a Path off of the specified location which can be
// absolute or relative.
func ParsePath(location string) (Path, error) {
	absLocation, err := filepath.Abs(location)
	if err != nil {
		return "", fmt.Errorf("error converting location %q to absolute: %w", location, err)
	}
	return Path(filepath.ToSlash(absLocation)), nil
}

// Path represents a filesystem location that is absolute and is in UNIX form.
type Path string

// String returns this Path as a string.
func (p Path) String() string {
	return string(p)
}

// Relative returns the Path without a leading separator.
func (p Path) Relative() Path {
	return Path(strings.TrimPrefix(string(p), "/"))
}

// CutSegment splits the path and returns the first segment in the path and the
// remaining path. If this path is not comprised of multiple segments then
// the returned path will be empty.
func (p Path) CutSegment() (string, Path) {
	before, after, ok := strings.Cut(string(p), "/")
	if !ok {
		return before, Path("")
	}
	return before, Path(after)
}
