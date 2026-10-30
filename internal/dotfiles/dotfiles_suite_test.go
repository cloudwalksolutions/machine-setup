package dotfiles_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDotfilesSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dotfiles Suite")
}
