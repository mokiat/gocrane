package location

import (
	"fmt"
	"strings"
)

const globPrefix = "*/"

// AppearsGlob checks whether the specified pattern is a candidate to be
// a glob. In general, that means it has the glob prefix, which is
// `*/` on Unix systems and `*\` on Windows systems.
// If this method returns true, this does not mean that the pattern
// is guaranteed to be successfully parsed as a glob.
func AppearsGlob(pattern string) bool {
	return strings.HasPrefix(pattern, globPrefix)
}

// Glob prepends the glob prefix to the specified pattern.
func Glob(pattern string) string {
	return fmt.Sprintf("%s%s", globPrefix, pattern)
}
