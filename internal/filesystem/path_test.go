package filesystem_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/filesystem"
)

var _ = Describe("Path", func() {

	Describe("ParsePath", func() {

		It("it should handle absolute paths", func() {
			stringPath := filepath.FromSlash("/tmp/../tmp/example")
			path, err := filesystem.ParsePath(stringPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal(filesystem.Path("/tmp/example")))
		})

		It("it should handle relative paths", func() {
			stringPath := filepath.FromSlash("./tmp/example")
			path, err := filesystem.ParsePath(stringPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(ContainSubstring("/tmp/example"))
		})

	})

})
