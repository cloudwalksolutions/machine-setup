package pkg_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPkgSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg Suite")
}
