package filesystem_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/filesystem"
)

var _ = Describe("WatchTree", func() {
	var tree *filesystem.WatchTree

	BeforeEach(func() {
		tree = filesystem.NewWatchTree()

		tree.Watch(filesystem.Path{
			"users",
		})
		tree.Ignore(filesystem.Path{
			"users", "max",
		})
		tree.Ignore(filesystem.Path{
			"users", "john", "documents",
		})
		tree.Watch(filesystem.Path{
			"users", "john", "documents", "memos",
		})
		tree.Ignore(filesystem.Path{
			"users", "john", "documents", "memos", "travel", "japan",
		})
		tree.Ignore(filesystem.Path{
			"users", "alice",
		})
	})

	Specify("root should not be watched", func() {
		Expect(tree.Navigate().ShouldWatch()).To(BeFalse())
	})

	Specify("watched segments should be watched", func() {
		Expect(tree.Navigate().Navigate("users").ShouldWatch()).To(BeTrue())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").ShouldWatch()).To(BeTrue())
	})

	Specify("ignored segments should not be watched", func() {
		Expect(tree.Navigate().Navigate("max").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("john").Navigate("documents").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("john").Navigate("documents").Navigate("memos").Navigate("travel").Navigate("japan").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("alice").ShouldWatch()).To(BeFalse())
	})

	Specify("segments after watched segments should be watched", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("jane").ShouldWatch()).To(BeTrue())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").Navigate("work").ShouldWatch()).To(BeTrue())
	})

	Specify("segments after ignored segments should not be watched", func() {
		Expect(tree.Navigate().Navigate("max").Navigate("documents").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("john").Navigate("documents").Navigate("videos").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("john").Navigate("documents").Navigate("memos").Navigate("travel").Navigate("japan").Navigate("tokyo").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("alice").Navigate("contacts").ShouldWatch()).To(BeFalse())
	})
})
