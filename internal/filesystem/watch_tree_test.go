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
		Expect(tree.Navigate().IsAccepted()).To(BeFalse())
	})

	Specify("segments of accepted paths should be accepted", func() {
		Expect(tree.Navigate().Navigate("users").IsAccepted()).To(BeTrue())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").IsAccepted()).To(BeTrue())
	})

	Specify("segments of rejected paths should be rejected", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("max").IsAccepted()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").IsAccepted()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").Navigate("travel").Navigate("japan").IsAccepted()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("alice").IsAccepted()).To(BeFalse())
	})

	Specify("segments after accepted path segments should be accepted", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("jane").IsAccepted()).To(BeTrue())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").Navigate("work").IsAccepted()).To(BeTrue())
	})

	Specify("segments after rejected path segments should be rejected", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("max").Navigate("documents").IsAccepted()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("videos").IsAccepted()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("john").Navigate("documents").Navigate("memos").Navigate("travel").Navigate("japan").Navigate("tokyo").IsAccepted()).To(BeFalse())
		Expect(tree.Navigate().Navigate("users").Navigate("alice").Navigate("contacts").IsAccepted()).To(BeFalse())
	})

	Specify("paths can be navigated in a single step", func() {
		Expect(tree.NavigatePath("/").IsAccepted()).To(BeFalse())
		Expect(tree.NavigatePath("").IsAccepted()).To(BeFalse())
		Expect(tree.NavigatePath("/users/john/documents/memos").IsAccepted()).To(BeTrue())
		Expect(tree.NavigatePath("/users/john/documents/memos/travel/japan").IsAccepted()).To(BeFalse())
	})

	Specify("segments matching rejected globs should be rejected", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("jane").Navigate("data_test.go").IsAccepted()).To(BeFalse())
	})

	Specify("segments matching accepted globs should be accepted", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("max").Navigate("some_important_items").IsAccepted()).To(BeTrue())
	})

	Specify("accepted globs supersede rejected globs", func() {
		Expect(tree.Navigate().Navigate("users").Navigate("max").Navigate("some_important_items_test.go").IsAccepted()).To(BeTrue())
	})
})
