package brew_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/pkg/brew"
)

var _ = Describe("TappedFormula.Install", func() {
	It("first taps the source, then installs the fully-qualified name", func() {
		var calls [][]string
		spy := func(args []string, _, _ io.Writer) error {
			calls = append(calls, args)
			return nil
		}

		err := brew.NewTappedFormula("terraform", "hashicorp/tap", spy).
			Install(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal([][]string{
			{"tap", "hashicorp/tap"},
			{"install", "hashicorp/tap/terraform"},
		}))
	})
})
