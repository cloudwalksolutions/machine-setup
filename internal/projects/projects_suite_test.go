package projects_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProjectsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "projects Suite")
}
