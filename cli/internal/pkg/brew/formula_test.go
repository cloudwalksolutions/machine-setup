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

var _ = Describe("Formula", func() {
	It("carries its name and description", func() {
		f := brew.NewFormula("yarn", "JavaScript package manager", nil)
		Expect(f.Name()).To(Equal("yarn"))
		Expect(f.Description()).To(Equal("JavaScript package manager"))
	})

	It("Install invokes the runner with [install <name>]", func() {
		var gotArgs []string
		spy := func(args []string, _, _ io.Writer) error {
			gotArgs = args
			return nil
		}

		err := brew.NewFormula("yarn", "", spy).Install(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(gotArgs).To(Equal([]string{"install", "yarn"}))
	})
})

var _ = Describe("Formula.Status", func() {
	It("reports the installed version from `brew list --versions`", func() {
		var gotArgs []string
		fake := func(args []string, stdout, _ io.Writer) error {
			gotArgs = args
			_, _ = io.WriteString(stdout, "yarn 1.22.22\n")
			return nil
		}

		status, version, err := brew.NewFormula("yarn", "", fake).Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotArgs).To(Equal([]string{"list", "--versions", "yarn"}))
		Expect(status).To(Equal(pkg.StatusUpToDate))
		Expect(version).To(Equal("1.22.22"))
	})

	It("falls back to the binary on PATH when brew does not list the formula", func() {
		brewMissing := func(_ []string, _, _ io.Writer) error { return errors.New("exit status 1") }
		f := brew.NewFormula("yarn", "", brewMissing)
		f.Probe = pkg.PathProbe{
			LookPath: func(file string) (string, error) { return "/usr/local/bin/" + file, nil },
			Run: func(cmd []string, stdout, _ io.Writer) error {
				Expect(cmd).To(Equal([]string{"/usr/local/bin/yarn", "--version"}))
				_, _ = io.WriteString(stdout, "1.22.22\n")
				return nil
			},
		}

		status, version, err := f.Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(pkg.StatusUpToDate))
		Expect(version).To(Equal("1.22.22"))
	})

	It("reports not installed when neither brew nor PATH has it", func() {
		brewMissing := func(_ []string, _, _ io.Writer) error { return errors.New("exit status 1") }
		f := brew.NewFormula("yarn", "", brewMissing)
		f.Probe = pkg.PathProbe{LookPath: func(string) (string, error) { return "", errors.New("not found") }}

		status, version, err := f.Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(pkg.StatusNotInstalled))
		Expect(version).To(BeEmpty())
	})
})
