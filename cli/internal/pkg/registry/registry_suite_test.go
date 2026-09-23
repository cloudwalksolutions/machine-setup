package registry_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRegistrySuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "registry Suite")
}
