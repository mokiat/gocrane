package filesystem

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Path represents a sequence of directories or files that represent
// an absolute file path on the filesystem.
type Path []string

// ParsePath constructs a Path off of the specified location which can be
// absolute or relative.
func ParsePath(location string) (Path, error) {
	absLocation, err := filepath.Abs(location)
	if err != nil {
		return nil, fmt.Errorf("error converting location %q to absolute: %w", location, err)
	}
	unixLocation := filepath.ToSlash(absLocation)
	trimmedLocation := strings.TrimPrefix(unixLocation, "/")
	trimmedLocation = strings.TrimSuffix(trimmedLocation, "/")
	return strings.Split(trimmedLocation, "/"), nil
}
