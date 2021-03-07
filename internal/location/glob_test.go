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

	Describe("Glob", func() {
		It("produces pattern with glob prefix", func() {
			glob := location.Glob("hello.txt")
			Expect(glob).To(Equal("*/hello.txt"))
		})
	})

})
