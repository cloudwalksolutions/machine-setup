package rvm_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/pkg/rvm"
)

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
