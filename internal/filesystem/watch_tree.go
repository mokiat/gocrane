package filesystem

import (
	"fmt"
	"path/filepath"

	"github.com/mokiat/gog/ds"
)

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

// RootPaths returns the top-most paths that are accepted.
func (t *FilterTree) RootPaths() []AbsolutePath {
	result := ds.NewList[string](0)
	for childName := range t.root.children {
		childNode, isChildAccepted := t.navigateAway(t.root, false, childName)
		if isChildAccepted {
			result.Add(childName)
		}
		t.findRoots(result, childName, childNode, isChildAccepted)
	}
	return result.Items()
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
func (t *FilterTree) AcceptPath(path AbsolutePath) {
	t.acceptRelativePath(t.root, path)
}

// RejectPath requests that the specified path be rejected.
func (t *FilterTree) RejectPath(path AbsolutePath) {
	t.rejectRelativePath(t.root, path)
}

// IsAccepted returns whether the specified path is allowed by this filter.
func (t *FilterTree) IsAccepted(path AbsolutePath) bool {
	var (
		current           = t.root
		isCurrentAccepted = false
	)
	childName, nextChildPath := CutPath(path)
	current, isCurrentAccepted = t.navigateAway(current, isCurrentAccepted, childName)
	for nextChildPath != "" {
		childName, nextChildPath = CutPath(nextChildPath)
		current, isCurrentAccepted = t.navigateAway(current, isCurrentAccepted, childName)
	}
	return isCurrentAccepted
}

func (t *FilterTree) acceptRelativePath(node *filterTreeNode, childPath string) {
	if len(childPath) == 0 {
		node.shouldAccept = true
		return
	}
	childName, nextChildPath := CutPath(childPath)
	childNode, ok := node.children[childName]
	if !ok {
		childNode = newFilterTreeNode()
		node.children[childName] = childNode
	}
	t.acceptRelativePath(childNode, nextChildPath)
}

func (t *FilterTree) rejectRelativePath(node *filterTreeNode, childPath string) {
	if len(childPath) == 0 {
		node.shouldReject = true
		return
	}
	childName, nextChildPath := CutPath(childPath)
	childNode, ok := node.children[childName]
	if !ok {
		childNode = newFilterTreeNode()
		node.children[childName] = childNode
	}
	t.rejectRelativePath(childNode, nextChildPath)
}

func (t *FilterTree) findRoots(result *ds.List[string], currentPath string, current *filterTreeNode, isCurrentAccepted bool) {
	for childName, childNode := range current.children {
		childPath := fmt.Sprintf("%s%s%s", currentPath, string(filepath.Separator), childName)
		_, isChildAccepted := t.navigateAway(current, isCurrentAccepted, childName)
		if isChildAccepted && !isCurrentAccepted {
			result.Add(childPath)
		}
		t.findRoots(result, childPath, childNode, isChildAccepted)
	}
}

func (t *FilterTree) navigateAway(current *filterTreeNode, isCurrentAccepted bool, childName string) (*filterTreeNode, bool) {
	var (
		childNode       *filterTreeNode
		childIsAccepted = isCurrentAccepted
	)
	// try and get a child node
	if current != nil {
		childNode = current.children[childName]
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
	if t.isSegmentPatternRejected(childName) {
		childIsAccepted = false
	}
	if t.isSegmentPatternAccepted(childName) {
		childIsAccepted = true
	}
	return childNode, childIsAccepted
}

func (t *FilterTree) isSegmentPatternAccepted(segment string) bool {
	for _, pattern := range t.acceptPatterns {
		if ok, err := filepath.Match(pattern, segment); err == nil && ok {
			return true
		}
	}
	return false
}

func (t *FilterTree) isSegmentPatternRejected(segment string) bool {
	for _, pattern := range t.rejectPatterns {
		if ok, err := filepath.Match(pattern, segment); err == nil && ok {
			return true
		}
	}
	return false
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
