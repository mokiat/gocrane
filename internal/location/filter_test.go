package location_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/location"
)

var _ = Describe("Filter", func() {
	var filter location.Filter

	Describe("GlobFilter", func() {
		BeforeEach(func() {
			glob := location.MustParseGlob(location.WithGlobPrefix("*.go"))
			filter = location.GlobFilter(glob)

		})

		DescribeTable("Match",
			func(p string, expected bool) {
				path := location.MustParsePath(filepath.FromSlash(p))
				Expect(filter.Match(path)).To(Equal(expected))
			},
			Entry("no segments match", "/a/b/c/d", false),
			Entry("one segment matches", "/a/b/c/d.go", true),
			Entry("multiple segments matche", "/a.go/b/c/d.go", true),
		)
	})

	Describe("PathFilter", func() {
		BeforeEach(func() {
			path := location.MustParsePath(filepath.FromSlash("/a/b/c"))
			filter = location.NewPathFilter(path)
		})

		DescribeTable("Match",
			func(p string, expected bool) {
				path := location.MustParsePath(filepath.FromSlash(p))
				Expect(filter.Match(path)).To(Equal(expected))
			},
			Entry("equal strings", "/a/b/c", true),
			Entry("different path", "/a/b/d", false),
			Entry("parent path", "/a/b", false),
			Entry("child path", "/a/b/c/d", true),
		)
	})

	Describe("OrFilter", func() {
		var (
			path              location.Path
			firstFilterMatch  bool
			firstFilter       location.Filter
			secondFilterMatch bool
			secondFilter      location.Filter
		)

		BeforeEach(func() {
			path = location.MustParsePath(filepath.FromSlash("/a/b/c"))

			firstFilterMatch = false
			firstFilter = location.FilterFunc(func(p location.Path) bool {
				Expect(p).To(Equal(path))
				return firstFilterMatch
			})

			secondFilterMatch = false
			secondFilter = location.FilterFunc(func(p location.Path) bool {
				Expect(p).To(Equal(path))
				return secondFilterMatch
			})

			filter = location.OrFilter(firstFilter, secondFilter)
		})

		DescribeTable("Match",
			func(firstMatch, secondMatch, expected bool) {
				firstFilterMatch = firstMatch
				secondFilterMatch = secondMatch
				Expect(filter.Match(path)).To(Equal(expected))
			},
			Entry("none matches", false, false, false),
			Entry("one matches", false, true, true),
			Entry("another matches", true, false, true),
			Entry("all match", true, true, true),
		)
	})

})
