package location

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Filter exposes methods for checking whether a Path matches
// a set of conditions.
type Filter interface {
	// Match checks whether the specified Path matches this filter.
	Match(path string) bool
}

// FilterFunc is a function that can performs a Match operation.
type FilterFunc func(path string) bool

// Match returns whether the specified path matches this filter.
func (f FilterFunc) Match(path string) bool {
	return f(path)
}

// GlobFilter creates a filter based on the specified Glob. A path matches
// this filter if any segment in it matches the Glob.
func GlobFilter(glob string) Filter {
	if !AppearsGlob(glob) {
		panic(fmt.Errorf("invalid glob: %q", glob))
	}
	pattern := strings.TrimPrefix(glob, globPrefix)
	return FilterFunc(func(path string) bool {
		segments := strings.Split(path, string(filepath.Separator))
		for _, segment := range segments {
			if segment == "" {
				continue
			}
			if ok, err := filepath.Match(pattern, segment); err == nil && ok {
				return true
			}
		}
		return false
	})
}

// PathFilter creates a filter based on the specified Path. Paths
// that are equal or children of the specified path will match this
// filter.
func PathFilter(filterPath string) Filter {
	return FilterFunc(func(path string) bool {
		return strings.HasPrefix(path, string(filterPath))
	})
}

// OrFilter returns a filter that matches a path should any of its
// sub-filters match that path.
func OrFilter(filters ...Filter) Filter {
	return FilterFunc(func(path string) bool {
		for _, filter := range filters {
			if filter.Match(path) {
				return true
			}
		}
		return false
	})
}

// AndFilter returns a filter that matches a path should all of its
// sub-filters match that path.
func AndFilter(filters ...Filter) Filter {
	return FilterFunc(func(path string) bool {
		for _, filter := range filters {
			if !filter.Match(path) {
				return false
			}
		}
		return true
	})
}

// NotFilter returns a filter that matches a path should its sub-filter
// not match it.
func NotFilter(filter Filter) Filter {
	return FilterFunc(func(path string) bool {
		return !filter.Match(path)
	})
}

func MatchAny(filter Filter, paths []string) bool {
	for _, path := range paths {
		if filter.Match(path) {
			return true
		}
	}
	return false
}
