package claudecode_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
	"tars/internal/pkg/claudecode"
)

var _ = Describe("claudecode.Installer", func() {
	var (
		path   string
		stdout *bytes.Buffer
		stderr *bytes.Buffer
	)

	BeforeEach(func() {
		tmp := GinkgoT().TempDir()
		path = filepath.Join(tmp, "claude")
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	Describe("Name()", func() {
		It("reports 'claude-code'", func() {
			installer := claudecode.NewInstaller(path, nil)
			Expect(installer.Name()).To(Equal("claude-code"))
		})
	})

	Describe("Install()", func() {
		It("is a no-op when the binary already exists", func() {
			Expect(os.WriteFile(path, []byte("fake binary"), 0o755)).To(Succeed())

			installer := claudecode.NewInstaller(path, func(_, _ io.Writer) error {
				panic("runner must not be called when binary exists")
			})

			Expect(installer.Install(stdout, stderr)).To(Succeed())
		})

		It("invokes the runner with standard writers when binary is missing", func() {
			var (
				gotStdout io.Writer
				gotStderr io.Writer
				calls     int
			)
			installer := claudecode.NewInstaller(path, func(o, e io.Writer) error {
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

	Describe("Status()", func() {
		It("reports NotInstalled when binary is missing", func() {
			installer := claudecode.NewInstaller(path, nil)
			status, version, err := installer.Status()
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(pkg.StatusNotInstalled))
			Expect(version).To(Equal(""))
		})

		It("reports UpToDate when binary exists", func() {
			Expect(os.WriteFile(path, []byte("fake binary"), 0o755)).To(Succeed())

			installer := claudecode.NewInstaller(path, nil)
			status, version, err := installer.Status()
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(pkg.StatusUpToDate))
			Expect(version).To(Equal("installed"))
		})
	})
})
