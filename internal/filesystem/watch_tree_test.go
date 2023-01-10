package filesystem_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/filesystem"
)

var _ = Describe("FilterTree", func() {
	var tree *filesystem.FilterTree

	BeforeEach(func() {
		tree = filesystem.NewFilterTree()

		tree.AcceptGlob(filesystem.Glob("*important*"))
		tree.RejectGlob(filesystem.Glob("*_test.go"))

		tree.AcceptPath("/users")
		tree.RejectPath("/users/max")
		tree.RejectPath("/users/john/documents")
		tree.AcceptPath("/users/john/documents/memos")
		tree.RejectPath("/users/john/documents/memos/travel/japan")
		tree.RejectPath("/users/alice")
	})

	Specify("root should not be accepted", func() {
		Expect(tree.IsAccepted("")).To(BeFalse())
		Expect(tree.IsAccepted("/")).To(BeFalse())
	})

	Specify("accepted paths should be accepted", func() {
		Expect(tree.IsAccepted("/users")).To(BeTrue())
		Expect(tree.IsAccepted("/users/john/documents/memos")).To(BeTrue())
	})

	Specify("rejected paths should be rejected", func() {
		Expect(tree.IsAccepted("/users/max")).To(BeFalse())
		Expect(tree.IsAccepted("/users/john/documents")).To(BeFalse())
		Expect(tree.IsAccepted("/users/john/documents/memos/travel/japan")).To(BeFalse())
		Expect(tree.IsAccepted("/users/alice")).To(BeFalse())
	})

	Specify("paths after accepted paths should be accepted", func() {
		Expect(tree.IsAccepted("/users/jane")).To(BeTrue())
		Expect(tree.IsAccepted("/users/john/documents/memos/work")).To(BeTrue())
	})

	Specify("paths after rejected path segments should be rejected", func() {
		Expect(tree.IsAccepted("/users/max/documents")).To(BeFalse())
		Expect(tree.IsAccepted("/users/john/documents/videos")).To(BeFalse())
		Expect(tree.IsAccepted("/users/john/documents/memos/travel/japan/tokyo")).To(BeFalse())
		Expect(tree.IsAccepted("/users/alice/contacts")).To(BeFalse())
	})

	Specify("paths matching rejected globs should be rejected", func() {
		Expect(tree.IsAccepted("/users/jane/data_test.go")).To(BeFalse())
	})

	Specify("paths matching accepted globs should be accepted", func() {
		Expect(tree.IsAccepted("/users/max/some_important_items")).To(BeTrue())
	})

	Specify("accepted globs supersede rejected globs", func() {
		Expect(tree.IsAccepted("/users/max/some_important_items_test.go")).To(BeTrue())
	})

	Specify("root paths can be extracted off of filtering rules", func() {
		paths := tree.RootPaths()
		Expect(paths).To(ConsistOf(
			"/users",
			"/users/john/documents/memos",
		))
	})
})
