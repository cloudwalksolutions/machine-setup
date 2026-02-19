package cmd

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSetupSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Machine Setup CLI Suite")
}
