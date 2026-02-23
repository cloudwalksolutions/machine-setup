package apt

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAptSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Apt Suite")
}
