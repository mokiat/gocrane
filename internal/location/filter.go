package location

// Filter exposes methods for checking whether a Path matches
// a set of conditions.
type Filter interface {
	// Match checks whether the specified Path matches this filter.
	Match(path Path) bool
}

// FilterFunc is a function that can performs a Match operation.
type FilterFunc func(path Path) bool

// Match returns whether the specified path matches this filter.
func (f FilterFunc) Match(path Path) bool {
	return f(path)
}

// GlobFilter creates a filter based on the specified Glob. A path matches
// this filter if any segment in it matches the Glob.
func GlobFilter(glob Glob) Filter {
	return FilterFunc(func(path Path) bool {
		for _, segment := range path {
			if glob.Match(segment) {
				return true
			}
		}
		return false
	})
}

// NewPathFilter creates a filter based on the specified Path. Paths
// that are equal or children of the specified path will match this
// filter.
func NewPathFilter(filterPath Path) Filter {
	return FilterFunc(func(path Path) bool {
		if len(path) < len(filterPath) {
			return false
		}
		for i, segment := range filterPath {
			if path[i] != segment {
				return false
			}
		}
		return true

	})
}

// OrFilter returns a filter that matches a path should any of its
// sub-filters match that path.
func OrFilter(filters ...Filter) Filter {
	return FilterFunc(func(path Path) bool {
		for _, filter := range filters {
			if filter.Match(path) {
				return true
			}
		}
		return false
	})
}
