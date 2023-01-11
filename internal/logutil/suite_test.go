package logutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLogutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Log Util Suite")
}
