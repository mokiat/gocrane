package logutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLogutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logutil Suite")
}
