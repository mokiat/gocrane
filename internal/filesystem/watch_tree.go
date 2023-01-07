package filesystem

func NewWatchTree() *WatchTree {
	return &WatchTree{
		root: newWatchNode(),
	}
}

type WatchTree struct {
	root *watchNode
}

func (t *WatchTree) Watch(path Path) {
	t.root.Watch(path)
}

func (t *WatchTree) Ignore(path Path) {
	t.root.Ignore(path)
}

func (t *WatchTree) Navigate() WatchCursor {
	return WatchCursor{
		node:        t.root,
		shouldWatch: t.root.shouldWatch,
	}
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

func (n *watchNode) Watch(childPath Path) {
	if len(childPath) == 0 {
		n.shouldWatch = true
		return
	}
	childNode, ok := n.children[childPath[0]]
	if !ok {
		childNode = newWatchNode()
		n.children[childPath[0]] = childNode
	}
	childNode.Watch(childPath[1:])
}

func (n *watchNode) Ignore(childPath Path) {
	if len(childPath) == 0 {
		n.shouldIgnore = true
		return
	}
	childNode, ok := n.children[childPath[0]]
	if !ok {
		childNode = newWatchNode()
		n.children[childPath[0]] = childNode
	}
	childNode.Ignore(childPath[1:])
}

type WatchCursor struct {
	node        *watchNode
	shouldWatch bool
}

func (c WatchCursor) ShouldWatch() bool {
	return c.shouldWatch
}

func (c WatchCursor) Navigate(segment string) WatchCursor {
	if c.node == nil {
		return c // can't navigate further, use last state as indicative
	}
	childNode := c.node.children[segment]
	if childNode == nil {
		return WatchCursor{
			node:        nil,
			shouldWatch: c.shouldWatch,
		}
	}
	shouldWatchChild := c.shouldWatch
	if childNode.shouldIgnore {
		shouldWatchChild = false
	}
	if childNode.shouldWatch {
		shouldWatchChild = true
	}
	return WatchCursor{
		node:        childNode,
		shouldWatch: shouldWatchChild,
	}
}
