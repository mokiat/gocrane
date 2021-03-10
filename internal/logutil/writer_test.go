package logutil_test

import (
	"io"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/mokiat/gocrane/internal/logutil"
)

var _ = Describe("Writer", func() {
	var (
		buffer *gbytes.Buffer
		writer io.Writer
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		logger := log.New(buffer, "test: ", log.Ldate)
		writer = logutil.ToWriter(logger)
	})

	It("writes out a partial line as a single line", func() {
		io.WriteString(writer, "first line")
		Expect(buffer).To(gbytes.Say("first line\n"))
	})

	It("writes out two lines as two lines", func() {
		io.WriteString(writer, "first line\n")
		io.WriteString(writer, "second line\n")
		Expect(buffer).To(gbytes.Say("first line\n"))
		Expect(buffer).To(gbytes.Say("second line\n"))
	})
})
