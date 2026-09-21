package brew_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg/brew"
)

var _ = Describe("TappedFormula.Install", func() {
	It("taps the repository, then installs the tapped formula", func() {
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

	It("stops when the tap step fails", func() {
		calls := 0
		spy := func(args []string, _, _ io.Writer) error {
			calls++
			return errors.New("tap failed")
		}

		err := brew.NewTappedFormula("terraform", "hashicorp/tap", spy).
			Install(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).To(MatchError(ContainSubstring("tap failed")))
		Expect(calls).To(Equal(1))
	})
})
