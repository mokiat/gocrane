package location_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/location"
)

var _ = Describe("Glob", func() {

	Describe("AppearsGlob", func() {
		It("rejects patterns with wrong prefix", func() {
			Expect(location.AppearsGlob("*|hello.txt")).To(BeFalse())
		})

		It("accepts patterns with valid prefix", func() {
			pattern := fmt.Sprintf("*%c%s", filepath.Separator, "*_test.go")
			Expect(location.AppearsGlob(pattern)).To(BeTrue())
		})
	})

	Describe("ParseGlob", func() {
		It("rejects patterns with wrong prefix", func() {
			_, err := location.ParseGlob("*|hello.txt")
			Expect(err).To(HaveOccurred())
		})

		It("rejects patterns with wrong filepath pattern", func() {
			pattern := fmt.Sprintf("*%c%s", filepath.Separator, "hello[a--b]")
			_, err := location.ParseGlob(pattern)
			Expect(err).To(HaveOccurred())
		})

		It("parses valid patterns", func() {
			pattern := fmt.Sprintf("*%c%s", filepath.Separator, "*_test.go")
			glob, err := location.ParseGlob(pattern)
			Expect(err).ToNot(HaveOccurred())
			Expect(glob.Match("hello_test.go")).To(BeTrue())
			Expect(glob.Match("hello.go")).To(BeFalse())
		})
	})

})
