package brew

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBrewSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brew Suite")
}
