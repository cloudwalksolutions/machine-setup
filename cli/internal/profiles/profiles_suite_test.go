package profiles_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProfilesSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "profiles Suite")
}
