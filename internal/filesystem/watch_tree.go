package filesystem

import "path/filepath"

// NewFilterTree creates a new empty FilterTree instance.
func NewFilterTree() *FilterTree {
	return &FilterTree{
		root: newFilterTreeNode(),
	}
}

// FilterTree is a data structure that can be used to mark specific filesystem
// paths as accepted and others as rejected. This can also be achieved through
// global glob patterns.
// The structure then provides a means through which one can test whether
// a given file path is accepted or rejected by the filter.
type FilterTree struct {

	// pattern related filtering
	acceptPatterns []string
	rejectPatterns []string

	// directory related filtering
	root *filterTreeNode
}

// AcceptGlob requests that sub-paths of a path segment that matches
// the specified glob should be accepted.
func (t *FilterTree) AcceptGlob(glob string) {
	t.acceptPatterns = append(t.acceptPatterns, Pattern(glob))
}

// RejectGlob requests that sub-paths of a path segment that matches
// the specified glob should not be accepted.
func (t *FilterTree) RejectGlob(glob string) {
	t.rejectPatterns = append(t.rejectPatterns, Pattern(glob))
}

// AcceptPath requests that the specified path be accepted.
func (t *FilterTree) AcceptPath(path Path) {
	t.acceptRelativePath(t.root, path.Relative())
}

// RejectPath requests that the specified path be rejected.
func (t *FilterTree) RejectPath(path Path) {
	t.rejectRelativePath(t.root, path.Relative())
}

// Navigate starts traversing the FilterTree beginning with the root
// for which a FilterTreeCursor is returned.
func (t *FilterTree) Navigate() FilterTreeCursor {
	return FilterTreeCursor{
		acceptPatterns: t.acceptPatterns,
		rejectPatterns: t.rejectPatterns,
		node:           t.root,
		isAccepted:     t.root.shouldAccept,
	}
}

// NavigatePath is a helper function that performs a sequence of Navigate
// calls using the specified Path as a guide.
func (t *FilterTree) NavigatePath(path Path) FilterTreeCursor {
	cursor := t.Navigate()
	path = path.Relative()
	for len(path) > 0 {
		segment, nextChildPath := path.CutSegment()
		cursor = cursor.Navigate(segment)
		path = nextChildPath
	}
	return cursor
}

func (t *FilterTree) acceptRelativePath(node *filterTreeNode, childPath Path) {
	if len(childPath) == 0 {
		node.shouldAccept = true
		return
	}
	childName, nextChildPath := childPath.CutSegment()
	childNode, ok := node.children[childName]
	if !ok {
		childNode = newFilterTreeNode()
		node.children[childName] = childNode
	}
	t.acceptRelativePath(childNode, nextChildPath)
}

func (t *FilterTree) rejectRelativePath(node *filterTreeNode, childPath Path) {
	if len(childPath) == 0 {
		node.shouldReject = true
		return
	}
	childName, nextChildPath := childPath.CutSegment()
	childNode, ok := node.children[childName]
	if !ok {
		childNode = newFilterTreeNode()
		node.children[childName] = childNode
	}
	t.rejectRelativePath(childNode, nextChildPath)
}

func newFilterTreeNode() *filterTreeNode {
	return &filterTreeNode{
		children: make(map[string]*filterTreeNode),
	}
}

type filterTreeNode struct {
	children     map[string]*filterTreeNode
	shouldAccept bool
	shouldReject bool
}

// FilterTreeCursor represents a particular path location in a FilterTree.
type FilterTreeCursor struct {
	acceptPatterns []string
	rejectPatterns []string
	node           *filterTreeNode
	isAccepted     bool
}

// IsAccepted returns whether the (sub-)path that is referenced by this
// FilterTreeCursor is accepted.
func (c FilterTreeCursor) IsAccepted() bool {
	return c.isAccepted
}

// Navigate returns a new FilterTreeCursor that is the result of advancing the
// existing cursor along the path using the specified segment.
//
// NOTE: The current cursor is not modified.
func (c FilterTreeCursor) Navigate(segment string) FilterTreeCursor {
	var (
		childNode       *filterTreeNode
		childIsAccepted = c.isAccepted
	)

	// try and get a child node
	if c.node != nil {
		childNode = c.node.children[segment]
	}

	// check path rules
	if childNode != nil {
		if childNode.shouldReject {
			childIsAccepted = false
		}
		if childNode.shouldAccept {
			childIsAccepted = true
		}
	}
	// check pattern rules
	if c.isSegmentPatternRejected(segment) {
		childIsAccepted = false
	}
	if c.isSegmentPatternAccepted(segment) {
		childIsAccepted = true
	}

	return FilterTreeCursor{
		acceptPatterns: c.acceptPatterns,
		rejectPatterns: c.rejectPatterns,
		node:           childNode,
		isAccepted:     childIsAccepted,
	}
}

func (c FilterTreeCursor) isSegmentPatternAccepted(segment string) bool {
	for _, pattern := range c.acceptPatterns {
		if ok, err := filepath.Match(pattern, segment); err == nil && ok {
			return true
		}
	}
	return false
}

func (c FilterTreeCursor) isSegmentPatternRejected(segment string) bool {
	for _, pattern := range c.rejectPatterns {
		if ok, err := filepath.Match(pattern, segment); err == nil && ok {
			return true
		}
	}
	return false
}
