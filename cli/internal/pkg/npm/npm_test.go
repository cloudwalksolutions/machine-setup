package npm_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
	"tars/internal/pkg/npm"
)

var _ = Describe("npm.Package", func() {
	Describe("Name()", func() {
		It("reports the package's display name", func() {
			pkg := npm.NewPackage("gemini-cli", "@google/gemini-cli", nil)
			Expect(pkg.Name()).To(Equal("gemini-cli"))
		})
	})

	Describe("Install()", func() {
		var (
			stdout *bytes.Buffer
			stderr *bytes.Buffer
		)

		BeforeEach(func() {
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
		})

		It("invokes the runner with 'install -g <packageName>' and streams output", func() {
			var (
				gotArgs   []string
				gotStdout io.Writer
				gotStderr io.Writer
				calls     int
			)

			runner := func(args []string, o, e io.Writer) error {
				calls++
				gotArgs = args
				gotStdout = o
				gotStderr = e
				return nil
			}

			pkg := npm.NewPackage("gemini-cli", "@google/gemini-cli", runner)
			Expect(pkg.Install(stdout, stderr)).To(Succeed())

			Expect(calls).To(Equal(1))
			Expect(gotArgs).To(Equal([]string{"install", "-g", "@google/gemini-cli"}))
			Expect(gotStdout).To(BeIdenticalTo(io.Writer(stdout)))
			Expect(gotStderr).To(BeIdenticalTo(io.Writer(stderr)))
		})

		It("propagates the runner's error", func() {
			expectedErr := errors.New("npm failed")
			runner := func(_ []string, _, _ io.Writer) error {
				return expectedErr
			}

			pkg := npm.NewPackage("gemini-cli", "@google/gemini-cli", runner)
			Expect(pkg.Install(stdout, stderr)).To(MatchError(expectedErr))
		})
	})

	Describe("Status()", func() {
		It("reads the global version from `npm ls -g --depth=0 --json <packageName>`", func() {
			var gotArgs []string
			runner := func(args []string, o, _ io.Writer) error {
				gotArgs = args
				_, _ = io.WriteString(o, `{"dependencies":{"@google/gemini-cli":{"version":"0.5.3"}}}`)
				return nil
			}

			status, version, err := npm.NewPackage("gemini-cli", "@google/gemini-cli", runner).Status()

			Expect(err).NotTo(HaveOccurred())
			Expect(gotArgs).To(Equal([]string{"ls", "-g", "--depth=0", "--json", "@google/gemini-cli"}))
			Expect(status).To(Equal(pkg.StatusUpToDate))
			Expect(version).To(Equal("0.5.3"))
		})

		It("reports not installed when the package is absent from the listing", func() {
			runner := func(_ []string, o, _ io.Writer) error {
				_, _ = io.WriteString(o, `{"dependencies":{}}`)
				return errors.New("exit status 1")
			}

			status, version, err := npm.NewPackage("gemini-cli", "@google/gemini-cli", runner).Status()

			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(pkg.StatusNotInstalled))
			Expect(version).To(BeEmpty())
		})
	})
})
