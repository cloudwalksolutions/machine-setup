package sessions_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSessionsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "sessions Suite")
}
