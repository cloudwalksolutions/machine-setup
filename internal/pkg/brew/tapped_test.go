package brew_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
	"tars/internal/pkg/brew"
)

var _ = Describe("TappedFormula", func() {
	It("carries its name and description", func() {
		t := brew.NewTappedFormula("terraform", "Infrastructure as code", "hashicorp/tap", nil)
		Expect(t.Name()).To(Equal("terraform"))
		Expect(t.Description()).To(Equal("Infrastructure as code"))
	})

	It("Install taps the repository, then installs the tapped formula", func() {
		var calls [][]string
		spy := func(args []string, _, _ io.Writer) error {
			calls = append(calls, args)
			return nil
		}

		err := brew.NewTappedFormula("terraform", "", "hashicorp/tap", spy).
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

		err := brew.NewTappedFormula("terraform", "", "hashicorp/tap", spy).
			Install(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).To(MatchError(ContainSubstring("tap failed")))
		Expect(calls).To(Equal(1))
	})
})

var _ = Describe("TappedFormula.Status", func() {
	It("reports the installed version from `brew list --versions <name>`", func() {
		var gotArgs []string
		fake := func(args []string, stdout, _ io.Writer) error {
			gotArgs = args
			_, _ = io.WriteString(stdout, "terraform 1.9.8\n")
			return nil
		}

		status, version, err := brew.NewTappedFormula("terraform", "", "hashicorp/tap", fake).Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotArgs).To(Equal([]string{"list", "--versions", "terraform"}))
		Expect(status).To(Equal(pkg.StatusUpToDate))
		Expect(version).To(Equal("1.9.8"))
	})
})
