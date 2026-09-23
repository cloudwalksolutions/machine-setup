package pkg_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
)

var _ = Describe("InstallStatus", func() {
	It("reports correct string representation", func() {
		Expect(pkg.StatusNotInstalled.String()).To(Equal("not installed"))
		Expect(pkg.StatusUpdateAvailable.String()).To(Equal("update available"))
		Expect(pkg.StatusUpToDate.String()).To(Equal("up to date"))
		Expect(pkg.InstallStatus(99).String()).To(Equal("unknown"))
	})
})

var _ = Describe("ScriptInstaller", func() {
	var (
		path   string
		stdout *bytes.Buffer
		stderr *bytes.Buffer
	)

	BeforeEach(func() {
		tmp := GinkgoT().TempDir()
		path = filepath.Join(tmp, "my-binary")
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	Describe("Name()", func() {
		It("reports the given name", func() {
			installer := pkg.NewScriptInstaller("my-tool", path, nil, nil)
			Expect(installer.Name()).To(Equal("my-tool"))
		})
	})

	Describe("Install()", func() {
		It("is a no-op when the checkPath already exists", func() {
			Expect(os.WriteFile(path, []byte("fake binary"), 0o755)).To(Succeed())

			installer := pkg.NewScriptInstaller("my-tool", path, []string{"curl"}, func(_ []string, _, _ io.Writer) error {
				panic("runner must not be called when binary exists")
			})

			Expect(installer.Install(stdout, stderr)).To(Succeed())
		})

		It("invokes the runner with the installCmd when checkPath is missing", func() {
			var (
				gotCmd    []string
				gotStdout io.Writer
				gotStderr io.Writer
				calls     int
			)
			installer := pkg.NewScriptInstaller("my-tool", path, []string{"install-step"}, func(cmd []string, o, e io.Writer) error {
				calls++
				gotCmd = cmd
				gotStdout, gotStderr = o, e
				return nil
			})

			Expect(installer.Install(stdout, stderr)).To(Succeed())

			Expect(calls).To(Equal(1))
			Expect(gotCmd).To(Equal([]string{"install-step"}))
			Expect(gotStdout).To(BeIdenticalTo(io.Writer(stdout)))
			Expect(gotStderr).To(BeIdenticalTo(io.Writer(stderr)))
		})
	})

	Describe("Status()", func() {
		It("reports StatusNotInstalled when checkPath is missing", func() {
			installer := pkg.NewScriptInstaller("my-tool", path, nil, nil)
			status, version, err := installer.Status()
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(pkg.StatusNotInstalled))
			Expect(version).To(Equal(""))
		})

		It("reports StatusUpToDate when checkPath exists", func() {
			Expect(os.WriteFile(path, []byte("fake binary"), 0o755)).To(Succeed())

			installer := pkg.NewScriptInstaller("my-tool", path, nil, nil)
			status, version, err := installer.Status()
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(pkg.StatusUpToDate))
			Expect(version).To(Equal("installed"))
		})
	})
})
