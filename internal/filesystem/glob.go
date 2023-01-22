package filesystem

import "strings"

const globPrefix = "*/"

// IsGlob checks whether the specified pattern is a candidate to be
// a glob. In general, that means it has the glob prefix, which is `*/` .
// If this method returns true, this does not mean that the glob pattern
// itself is valid.
func IsGlob(pattern string) bool {
	return strings.HasPrefix(pattern, globPrefix)
}

// Glob prepends the glob prefix to the specified pattern.
func Glob(pattern string) string {
	return globPrefix + pattern
}

// Pattern returns the glob pattern from the specified glob string.
func Pattern(glob string) string {
	return strings.TrimPrefix(glob, globPrefix)
}
