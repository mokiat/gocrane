package location_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/location"
)

var _ = Describe("Path", func() {
	var fixture Fixture

	BeforeEach(func() {
		Expect(fixture.Create()).To(Succeed())
	})

	AfterEach(func() {
		Expect(fixture.Delete()).To(Succeed())
	})

	Describe("ParsePath", func() {
		It("should succeed on valid path", func() {
			Expect(fixture.CreateDir("/hello/world")).To(Succeed())

			parsedRoot, err := location.ParsePath(fixture.Root())
			Expect(err).ToNot(HaveOccurred())

			parsedPath, err := location.ParsePath(fixture.Path("./a/../hello/world"))
			Expect(err).ToNot(HaveOccurred())

			Expect(len(parsedPath)).To(Equal(len(parsedRoot) + 2))
			for i, segment := range parsedRoot {
				Expect(parsedPath[i]).To(Equal(segment))
			}

			offset := len(parsedRoot)
			Expect(parsedPath[offset+0]).To(Equal("hello"))
			Expect(parsedPath[offset+1]).To(Equal("world"))
		})
	})

})
