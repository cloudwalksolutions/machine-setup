package rvm_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
	"tars/internal/pkg/rvm"
)

var _ = Describe("rvm.Installer.Status", func() {
	It("reports the version from <dir>/VERSION when the rvm dir exists", func() {
		dir := filepath.Join(GinkgoT().TempDir(), ".rvm")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.29.12\n"), 0o644)).To(Succeed())

		status, detail, err := rvm.NewInstaller(dir, nil).Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(pkg.StatusUpToDate))
		Expect(detail).To(Equal("1.29.12"))
	})

	It("reports not installed when the rvm dir is missing", func() {
		status, detail, err := rvm.NewInstaller(filepath.Join(GinkgoT().TempDir(), ".rvm"), nil).Status()

		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(pkg.StatusNotInstalled))
		Expect(detail).To(BeEmpty())
	})
})

var _ = Describe("rvm.Installer.Install", func() {
	var (
		dir    string
		stdout *bytes.Buffer
		stderr *bytes.Buffer
	)

	BeforeEach(func() {
		tmp := GinkgoT().TempDir()
		dir = filepath.Join(tmp, ".rvm")
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	It("is a no-op when the rvm dir already exists", func() {
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())

		installer := rvm.NewInstaller(dir, func(_, _ io.Writer) error {
			panic("runner must not be called when dir exists")
		})

		Expect(installer.Install(stdout, stderr)).To(Succeed())
	})

	It("invokes the runner with the caller's writers when Dir is missing", func() {
		var (
			gotStdout io.Writer
			gotStderr io.Writer
			calls     int
		)
		installer := rvm.NewInstaller(dir, func(o, e io.Writer) error {
			calls++
			gotStdout, gotStderr = o, e
			return nil
		})

		Expect(installer.Install(stdout, stderr)).To(Succeed())

		Expect(calls).To(Equal(1))
		Expect(gotStdout).To(BeIdenticalTo(io.Writer(stdout)))
		Expect(gotStderr).To(BeIdenticalTo(io.Writer(stderr)))
	})
})
