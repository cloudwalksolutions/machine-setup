package assets_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAssetsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "assets Suite")
}
