package fsutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFsutilSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "fsutil Suite")
}
