package location

import (
	"fmt"
	"path/filepath"
	"strings"
)

var globPrefix string

func init() {
	globPrefix = fmt.Sprintf("*%c", filepath.Separator)
}

// AppearsGlob checks whether the specified pattern is a candidate to be
// a glob. In general, that means it has the glob prefix, which is
// `*/` on Unix systems and `*\` on Windows systems.
// If this method returns true, this does not mean that the pattern
// is guaranteed to be successfully parsed as a glob.
func AppearsGlob(pattern string) bool {
	return strings.HasPrefix(pattern, globPrefix)
}

// ParseGlob attempts to parse the specified pattern as a location Glob.
// One prerequisite is that it has to have the glob prefix. Check the
// AppearsGlob function for more information.
// Once the prefix is trimmed, the pattern must comply with the rules
// specified in `filepath.Match`. Keep in mind that patterns are evaluated
// only against individual Path segments.
func ParseGlob(pattern string) (Glob, error) {
	if !AppearsGlob(pattern) {
		return Glob{}, fmt.Errorf("pattern lacks necessary prefix")
	}
	trimmedPattern := strings.TrimPrefix(pattern, globPrefix)
	if _, err := filepath.Match(trimmedPattern, ""); err != nil {
		return Glob{}, fmt.Errorf("specified pattern is not valid: %w", err)
	}
	return Glob{
		pattern: trimmedPattern,
	}, nil
}

// Glob represents a pattern that can be checked against a path segment.
type Glob struct {
	pattern string
}

// Match returns whether the specified Path segment matches the Glob
// pattern.
func (g Glob) Match(segment string) bool {
	match, _ := filepath.Match(g.pattern, segment)
	return match
}
