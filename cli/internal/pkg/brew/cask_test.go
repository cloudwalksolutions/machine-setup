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

var _ = Describe("Cask", func() {
	It("carries its name and description", func() {
		c := brew.NewCask("gcloud-cli", "Google Cloud SDK", nil)
		Expect(c.Name()).To(Equal("gcloud-cli"))
		Expect(c.Description()).To(Equal("Google Cloud SDK"))
	})

	It("Install invokes the runner with [install --cask <name>]", func() {
		var gotArgs []string
		spy := func(args []string, _, _ io.Writer) error {
			gotArgs = args
			return nil
		}

		err := brew.NewCask("rustup", "", spy).Install(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(gotArgs).To(Equal([]string{"install", "--cask", "rustup"}))
	})
})

var _ = Describe("Cask.Status", func() {
	It("reports the installed version from `brew list --versions --cask`", func() {
		var gotArgs []string
		fake := func(args []string, stdout, _ io.Writer) error {
			gotArgs = args
			_, _ = io.WriteString(stdout, "gcloud-cli 540.0.0\n")
			return nil
		}

		status, version, err := brew.NewCask("gcloud-cli", "", fake).Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(gotArgs).To(Equal([]string{"list", "--versions", "--cask", "gcloud-cli"}))
		Expect(status).To(Equal(pkg.StatusUpToDate))
		Expect(version).To(Equal("540.0.0"))
	})

	It("falls back to its Binary on PATH when brew does not list the cask", func() {
		brewMissing := func(_ []string, _, _ io.Writer) error { return errors.New("exit status 1") }
		c := brew.NewCask("gcloud-cli", "", brewMissing)
		c.Binary = "gcloud"
		c.Probe = pkg.PathProbe{
			LookPath: func(file string) (string, error) { return "/opt/homebrew/bin/" + file, nil },
			Run: func(cmd []string, stdout, _ io.Writer) error {
				Expect(cmd).To(Equal([]string{"/opt/homebrew/bin/gcloud", "--version"}))
				_, _ = io.WriteString(stdout, "Google Cloud SDK 585.0.0\n")
				return nil
			},
		}

		status, version, err := c.Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(pkg.StatusUpToDate))
		Expect(version).To(Equal("585.0.0"))
	})
})
