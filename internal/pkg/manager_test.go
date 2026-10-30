package pkg_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

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

type fakeInstallable struct {
	name string
	log  *[]string
	err  error
}

func (f fakeInstallable) Name() string        { return f.name }
func (f fakeInstallable) Description() string { return "fake " + f.name }
func (f fakeInstallable) Install(_, _ io.Writer) error {
	*f.log = append(*f.log, "install "+f.name)
	return f.err
}
func (f fakeInstallable) Status() (pkg.InstallStatus, string, error) {
	return pkg.StatusUpToDate, "9.9.9", nil
}

var _ = Describe("WithPostInstall", func() {
	var log []string
	steps := [][]string{{"rustup", "install", "stable"}, {"rustup", "default", "stable"}}
	run := func(cmd []string, _, _ io.Writer) error {
		log = append(log, strings.Join(cmd, " "))
		return nil
	}

	BeforeEach(func() { log = nil })

	It("runs the steps in order after a successful install and delegates Name and Status", func() {
		wrapped := pkg.WithPostInstall(fakeInstallable{name: "rustup", log: &log}, steps, run)

		Expect(wrapped.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())

		Expect(log).To(Equal([]string{"install rustup", "rustup install stable", "rustup default stable"}))
		Expect(wrapped.Name()).To(Equal("rustup"))
		_, version, _ := wrapped.Status()
		Expect(version).To(Equal("9.9.9"))
	})

	It("skips the steps when the install fails", func() {
		wrapped := pkg.WithPostInstall(fakeInstallable{name: "rustup", log: &log, err: errors.New("brew failed")}, steps, run)

		Expect(wrapped.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(MatchError("brew failed"))

		Expect(log).To(Equal([]string{"install rustup"}))
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

	Describe("Name() and Description()", func() {
		It("report what was given", func() {
			installer := pkg.NewScriptInstaller("my-tool", "does things", path, nil, nil)
			Expect(installer.Name()).To(Equal("my-tool"))
			Expect(installer.Description()).To(Equal("does things"))
		})
	})

	Describe("Install()", func() {
		It("is a no-op when the checkPath already exists", func() {
			Expect(os.WriteFile(path, []byte("fake binary"), 0o755)).To(Succeed())

			installer := pkg.NewScriptInstaller("my-tool", "", path, []string{"curl"}, func(_ []string, _, _ io.Writer) error {
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
			installer := pkg.NewScriptInstaller("my-tool", "", path, []string{"install-step"}, func(cmd []string, o, e io.Writer) error {
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
			installer := pkg.NewScriptInstaller("my-tool", "", path, nil, nil)
			status, version, err := installer.Status()
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(pkg.StatusNotInstalled))
			Expect(version).To(Equal(""))
		})

		It("reports the first word of `<checkPath> --version` when checkPath exists", func() {
			Expect(os.WriteFile(path, []byte("fake binary"), 0o755)).To(Succeed())
			var gotCmd []string
			installer := pkg.NewScriptInstaller("my-tool", "", path, nil, func(cmd []string, o, _ io.Writer) error {
				gotCmd = cmd
				_, _ = io.WriteString(o, "2.1.0 (Claude Code)\n")
				return nil
			})

			status, version, err := installer.Status()

			Expect(err).NotTo(HaveOccurred())
			Expect(gotCmd).To(Equal([]string{path, "--version"}))
			Expect(status).To(Equal(pkg.StatusUpToDate))
			Expect(version).To(Equal("2.1.0"))
		})
	})
})
