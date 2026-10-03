package pkg_test

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
)

var _ = Describe("PathProbe.Status", func() {
	It("reports the first version token printed by `<binary> --version` when it is on PATH", func() {
		var gotCmd []string
		probe := pkg.PathProbe{
			LookPath: func(file string) (string, error) { return "/opt/homebrew/bin/" + file, nil },
			Run: func(cmd []string, stdout, _ io.Writer) error {
				gotCmd = cmd
				_, _ = io.WriteString(stdout, "Google Cloud SDK 585.0.0\nbq 2.1.0\n")
				return nil
			},
		}

		status, version, err := probe.Status("gcloud")

		Expect(err).NotTo(HaveOccurred())
		Expect(gotCmd).To(Equal([]string{"/opt/homebrew/bin/gcloud", "--version"}))
		Expect(status).To(Equal(pkg.StatusUpToDate))
		Expect(version).To(Equal("585.0.0"))
	})

	It("reports installed with an unknown version when --version fails", func() {
		probe := pkg.PathProbe{
			LookPath: func(file string) (string, error) { return "/usr/local/bin/" + file, nil },
			Run:      func([]string, io.Writer, io.Writer) error { return errors.New("exit status 2") },
		}

		status, version, err := probe.Status("k9s")

		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(pkg.StatusUpToDate))
		Expect(version).To(BeEmpty())
	})

	It("reports not installed when the binary is not on PATH", func() {
		probe := pkg.PathProbe{
			LookPath: func(string) (string, error) { return "", errors.New("not found") },
			Run:      func([]string, io.Writer, io.Writer) error { panic("must not run") },
		}

		status, version, err := probe.Status("gcloud")

		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(pkg.StatusNotInstalled))
		Expect(version).To(BeEmpty())
	})
})
