package brew_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/pkg/brew"
)

var _ = Describe("Formula.Install", func() {
	It("invokes the runner with [install <name>]", func() {
		var gotArgs []string
		spy := func(args []string, _, _ io.Writer) error {
			gotArgs = args
			return nil
		}

		err := brew.NewFormula("yarn", spy).Install(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(gotArgs).To(Equal([]string{"install", "yarn"}))
	})
})
