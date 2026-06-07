package shell_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/shell"
)

var _ = Describe("OhMyZshInstaller.Install", func() {
	var (
		ohmyzshHome string
		stdout      *bytes.Buffer
		stderr      *bytes.Buffer
	)

	BeforeEach(func() {
		tmp := GinkgoT().TempDir()
		ohmyzshHome = filepath.Join(tmp, ".oh-my-zsh")
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	})

	It("is a no-op when the oh-my-zsh dir already exists", func() {
		Expect(os.MkdirAll(ohmyzshHome, 0o755)).To(Succeed())

		installer := shell.OhMyZshInstaller{
			Dir:    ohmyzshHome,
			Runner: func(_, _ io.Writer) error { panic("runner must not be called") },
			Stdout: stdout,
			Stderr: stderr,
		}
		Expect(installer.Install()).To(Succeed())
	})

	It("invokes the runner with the configured writers when Dir is missing", func() {
		var (
			gotStdout io.Writer
			gotStderr io.Writer
			calls     int
		)
		installer := shell.OhMyZshInstaller{
			Dir: ohmyzshHome,
			Runner: func(o, e io.Writer) error {
				calls++
				gotStdout, gotStderr = o, e
				return nil
			},
			Stdout: stdout,
			Stderr: stderr,
		}
		Expect(installer.Install()).To(Succeed())
		Expect(calls).To(Equal(1))
		Expect(gotStdout).To(BeIdenticalTo(io.Writer(stdout)))
		Expect(gotStderr).To(BeIdenticalTo(io.Writer(stderr)))
	})
})
