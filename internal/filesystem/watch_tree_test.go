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

		tree.WatchGlob(filesystem.Glob("*important*"))
		tree.IgnoreGlob(filesystem.Glob("*_test.go"))

		tree.Watch("/users")
		tree.Ignore("/users/max")
		tree.Ignore("/users/john/documents")
		tree.Watch("/users/john/documents/memos")
		tree.Ignore("/users/john/documents/memos/travel/japan")
		tree.Ignore("/users/alice")
	})

	Specify("root should not be watched", func() {
		Expect(tree.Navigate().ShouldWatch()).To(BeFalse())
	})

	Specify("watched segments should be watched", func() {
		Expect(tree.Navigate().Navigate("users").ShouldWatch()).To(BeTrue())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").ShouldWatch()).To(BeTrue())
	})

	Specify("ignored segments should not be watched", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("max").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").Navigate("travel").Navigate("japan").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("alice").ShouldWatch()).To(BeFalse())
	})

	Specify("segments after watched segments should be watched", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("jane").ShouldWatch()).To(BeTrue())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").Navigate("work").ShouldWatch()).To(BeTrue())
	})

	Specify("segments after ignored segments should not be watched", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("max").Navigate("documents").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("videos").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").Navigate("travel").Navigate("japan").Navigate("tokyo").ShouldWatch()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("alice").Navigate("contacts").ShouldWatch()).To(BeFalse())
	})

	Specify("paths can be navigated in a single step", func() {
		Expect(tree.NavigatePath("/").ShouldWatch()).To(BeFalse())
		Expect(tree.NavigatePath("").ShouldWatch()).To(BeFalse())
		Expect(tree.NavigatePath("/users/john/documents/memos").ShouldWatch()).To(BeTrue())
		Expect(tree.NavigatePath("/users/john/documents/memos/travel/japan").ShouldWatch()).To(BeFalse())
	})

	Specify("segments matching ignored globs should not be watched", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("jane").Navigate("data_test.go").ShouldWatch()).To(BeFalse())
	})

	Specify("segments matching watched globs should be watched", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("max").Navigate("some_important_items").ShouldWatch()).To(BeTrue())
	})

	Specify("watched globs supersede ignored globs", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("max").Navigate("some_important_items_test.go").ShouldWatch()).To(BeTrue())
	})
})
