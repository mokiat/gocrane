package filesystem

import "path/filepath"

// NewWatchTree creates a new empty WatchTree instance.
func NewWatchTree() *WatchTree {
	return &WatchTree{
		root: newWatchNode(),
	}
}

// WatchTree is a data structure that can be used to mark specific filesystem
// paths as watchable or to be ignored. In addition, it provides mechanisms
// to control that globally through generic glob patterns.
//
// NOTE: This structure does not actually perform the watching. It just acts
// as a form of filter.
type WatchTree struct {

	// pattern related filtering
	watchPatterns  []string
	ignorePatterns []string

	// directory related filtering
	root *watchNode
}

// WatchGlob requests that sub-paths of a path segment that matches
// the specified glob should be watched.
func (t *WatchTree) WatchGlob(glob string) {
	t.watchPatterns = append(t.watchPatterns, Pattern(glob))
}

// IgnoreGlob requests that sub-paths of a path segment that matches
// the specified glob should not be watched.
func (t *WatchTree) IgnoreGlob(glob string) {
	t.ignorePatterns = append(t.ignorePatterns, Pattern(glob))
}

// Watch requests that the specified path be watched.
func (t *WatchTree) Watch(path Path) {
	t.watchRelativePath(t.root, path.Relative())
}

// Ignore requests that the specified path be ignored.
func (t *WatchTree) Ignore(path Path) {
	t.ignoreRelativePath(t.root, path.Relative())
}

// Navigate starts traversing the WatchTree beginning with the root
// for which a WatchCursor is returned.
func (t *WatchTree) Navigate() WatchCursor {
	return WatchCursor{
		watchPatterns:  t.watchPatterns,
		ignorePatterns: t.ignorePatterns,
		node:           t.root,
		shouldWatch:    t.root.shouldWatch,
	}
}

// NavigatePath is a helper function that performs a sequence of Navigate
// calls using the specified Path as a guide.
func (t *WatchTree) NavigatePath(path Path) WatchCursor {
	cursor := t.Navigate()
	path = path.Relative()
	for len(path) > 0 {
		segment, nextChildPath := path.CutSegment()
		cursor = cursor.Navigate(segment)
		path = nextChildPath
	}
	return cursor
}

func (t *WatchTree) watchRelativePath(node *watchNode, childPath Path) {
	if len(childPath) == 0 {
		node.shouldWatch = true
		return
	}
	childName, nextChildPath := childPath.CutSegment()
	childNode, ok := node.children[childName]
	if !ok {
		childNode = newWatchNode()
		node.children[childName] = childNode
	}
	t.watchRelativePath(childNode, nextChildPath)
}

func (t *WatchTree) ignoreRelativePath(node *watchNode, childPath Path) {
	if len(childPath) == 0 {
		node.shouldIgnore = true
		return
	}
	childName, nextChildPath := childPath.CutSegment()
	childNode, ok := node.children[childName]
	if !ok {
		childNode = newWatchNode()
		node.children[childName] = childNode
	}
	t.ignoreRelativePath(childNode, nextChildPath)
}

func newWatchNode() *watchNode {
	return &watchNode{
		children: make(map[string]*watchNode),
	}
}

type watchNode struct {
	children     map[string]*watchNode
	shouldWatch  bool
	shouldIgnore bool
}

// WatchCursor represents a particular path location in a WatchTree.
type WatchCursor struct {
	watchPatterns  []string
	ignorePatterns []string
	node           *watchNode
	shouldWatch    bool
}

// ShouldWatch returns whether the (sub-)path that is referenced by this
// WatchCursor should be watched.
func (c WatchCursor) ShouldWatch() bool {
	return c.shouldWatch
}

// Navigate returns a new WatchCursor that is the result of advancing the
// existing cursor along the path using the specified segment.
//
// NOTE: The current cursor is not modified.
func (c WatchCursor) Navigate(segment string) WatchCursor {
	var (
		childNode        *watchNode
		shouldWatchChild = c.shouldWatch
	)

	// try and get a child node
	if c.node != nil {
		childNode = c.node.children[segment]
	}

	// check path rules
	if childNode != nil {
		if childNode.shouldIgnore {
			shouldWatchChild = false
		}
		if childNode.shouldWatch {
			shouldWatchChild = true
		}
	}
	// check pattern rules
	if c.isSegmentPatternIgnored(segment) {
		shouldWatchChild = false
	}
	if c.isSegmentPatternWatched(segment) {
		shouldWatchChild = true
	}

	return WatchCursor{
		watchPatterns:  c.watchPatterns,
		ignorePatterns: c.ignorePatterns,
		node:           childNode,
		shouldWatch:    shouldWatchChild,
	}
}

func (c WatchCursor) isSegmentPatternWatched(segment string) bool {
	for _, pattern := range c.watchPatterns {
		if ok, err := filepath.Match(pattern, segment); err == nil && ok {
			return true
		}
	}
	return false
}

func (c WatchCursor) isSegmentPatternIgnored(segment string) bool {
	for _, pattern := range c.ignorePatterns {
		if ok, err := filepath.Match(pattern, segment); err == nil && ok {
			return true
		}
	}
	return false
}
