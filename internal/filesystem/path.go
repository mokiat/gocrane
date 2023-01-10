package filesystem

import (
	"path/filepath"
	"strings"
)

// ToAbsolutePath converts the specified path to absolute.
func ToAbsolutePath(path string) (AbsolutePath, error) {
	absLocation, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absLocation, nil
}

// AbsolutePath is a string that holds an absolute path.
type AbsolutePath = string

// CutPath tries to extract the first segment from the path.
// If the path does not contain a separator, then the returned
// remainder is empty.
func CutPath(path string) (segment string, remainder string) {
	first, rest, _ := strings.Cut(path, string(filepath.Separator))
	return first, rest
}
