package filesystem_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/filesystem"
)

var _ = Describe("Glob", func() {

	Describe("IsGlob", func() {
		It("rejects patterns with wrong prefix", func() {
			Expect(filesystem.IsGlob("*|hello.txt")).To(BeFalse())
		})

		It("accepts patterns with valid prefix", func() {
			Expect(filesystem.IsGlob("*/*_test.go")).To(BeTrue())
		})
	})

	Describe("Glob", func() {
		It("produces pattern with glob prefix", func() {
			Expect(filesystem.Glob("hello.txt")).To(Equal("*/hello.txt"))
		})
	})

	Describe("Pattern", func() {
		It("extracts the glob pattern", func() {
			Expect(filesystem.Pattern("*/hello.txt")).To(Equal("hello.txt"))
		})
	})

})
