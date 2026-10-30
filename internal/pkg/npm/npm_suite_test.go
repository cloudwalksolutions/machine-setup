package npm_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNpmSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "npm Suite")
}
