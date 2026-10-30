package rvm_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRvmSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "rvm Suite")
}
