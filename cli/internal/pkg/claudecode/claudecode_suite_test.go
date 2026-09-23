package claudecode_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClaudecodeSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "claudecode Suite")
}
