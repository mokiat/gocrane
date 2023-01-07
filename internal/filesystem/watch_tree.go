package filesystem

import "path/filepath"

func NewWatchTree() *WatchTree {
	return &WatchTree{
		root: newWatchNode(),
	}
}

type WatchTree struct {
	// pattern related filtering
	watchPatterns  []string
	ignorePatterns []string

	// directory related filtering
	root *watchNode
}

func (t *WatchTree) WatchGlob(glob string) {
	t.watchPatterns = append(t.watchPatterns, Pattern(glob))
}

func (t *WatchTree) IgnoreGlob(glob string) {
	t.ignorePatterns = append(t.ignorePatterns, Pattern(glob))
}

func (t *WatchTree) Watch(path Path) {
	t.watchRelativePath(t.root, path)
}

func (t *WatchTree) Ignore(path Path) {
	t.ignoreRelativePath(t.root, path)
}

func (t *WatchTree) Navigate() WatchCursor {
	return WatchCursor{
		watchPatterns:  t.watchPatterns,
		ignorePatterns: t.ignorePatterns,
		node:           t.root,
		shouldWatch:    t.root.shouldWatch,
	}
}

func (t *WatchTree) NavigatePath(path Path) WatchCursor {
	cursor := t.Navigate()
	for _, segment := range path {
		cursor = cursor.Navigate(segment)
	}
	return cursor
}

func (t *WatchTree) watchRelativePath(node *watchNode, childPath Path) {
	if len(childPath) == 0 {
		node.shouldWatch = true
		return
	}
	childName := childPath[0]
	childNode, ok := node.children[childName]
	if !ok {
		childNode = newWatchNode()
		node.children[childName] = childNode
	}
	t.watchRelativePath(childNode, childPath[1:])
}

func (t *WatchTree) ignoreRelativePath(node *watchNode, childPath Path) {
	if len(childPath) == 0 {
		node.shouldIgnore = true
		return
	}
	childName := childPath[0]
	childNode, ok := node.children[childName]
	if !ok {
		childNode = newWatchNode()
		node.children[childName] = childNode
	}
	t.ignoreRelativePath(childNode, childPath[1:])
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

type WatchCursor struct {
	watchPatterns  []string
	ignorePatterns []string
	node           *watchNode
	shouldWatch    bool
}

func (c WatchCursor) ShouldWatch() bool {
	return c.shouldWatch
}

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
