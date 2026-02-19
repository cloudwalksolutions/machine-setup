package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

// ExecutorFixture captures cobra command output and manages test config paths.
type ExecutorFixture struct {
	TmpDir     string
	ConfigPath string
	Stdout     *bytes.Buffer
	Stderr     *bytes.Buffer
}

func NewExecutorFixture() *ExecutorFixture {
	return &ExecutorFixture{}
}

func (f *ExecutorFixture) Setup() {
	var err error
	f.TmpDir, err = os.MkdirTemp("", "machine-setup-test-*")
	Expect(err).NotTo(HaveOccurred())

	f.ConfigPath = filepath.Join(f.TmpDir, "config.yaml")

	// Redirect config path and suppress TUI form via env vars.
	os.Setenv("MACHINE_SETUP_CONFIG_PATH", f.ConfigPath)
	os.Setenv("MACHINE_SETUP_NO_FORM", "1")
}

func (f *ExecutorFixture) Teardown() {
	os.RemoveAll(f.TmpDir)
	os.Unsetenv("MACHINE_SETUP_CONFIG_PATH")
	os.Unsetenv("MACHINE_SETUP_NO_FORM")
}

// RunSetup executes the setup command via cobra, capturing stdout and stderr.
func (f *ExecutorFixture) RunSetup(extraArgs ...string) error {
	f.Stdout = new(bytes.Buffer)
	f.Stderr = new(bytes.Buffer)

	cfgFile = "" // reset persistent flag state between runs
	rootCmd.SetOut(f.Stdout)
	rootCmd.SetErr(f.Stderr)
	rootCmd.SetArgs(append([]string{"setup"}, extraArgs...))

	return rootCmd.Execute()
}

// ReadConfig parses the YAML config file written by the last RunSetup call.
func (f *ExecutorFixture) ReadConfig() map[string]interface{} {
	data, err := os.ReadFile(f.ConfigPath)
	Expect(err).NotTo(HaveOccurred())

	var result map[string]interface{}
	Expect(yaml.Unmarshal(data, &result)).To(Succeed())
	return result
}

var _ = Describe("setup command", func() {
	var fixture *ExecutorFixture

	BeforeEach(func() {
		fixture = NewExecutorFixture()
		fixture.Setup()
	})

	AfterEach(func() {
		fixture.Teardown()
	})

	Describe("config file creation", func() {
		It("creates the config file at the expected path", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(fixture.ConfigPath).To(BeAnExistingFile())
		})

		It("writes valid YAML to the config file", func() {
			Expect(fixture.RunSetup()).To(Succeed())

			data, err := os.ReadFile(fixture.ConfigPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).NotTo(BeEmpty())
		})

		It("includes the architecture key in the config", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(fixture.ReadConfig()).To(HaveKey("architecture"))
		})

		It("writes the correct architecture for the current machine", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(fixture.ReadConfig()["architecture"]).To(Equal(runtime.GOARCH))
		})

		It("initializes sources as an empty list", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			cfg := fixture.ReadConfig()
			if sources, ok := cfg["sources"]; ok && sources != nil {
				Expect(sources).To(BeEmpty())
			}
		})

		It("initializes packages as an empty list", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			cfg := fixture.ReadConfig()
			if packages, ok := cfg["packages"]; ok && packages != nil {
				Expect(packages).To(BeEmpty())
			}
		})

		It("initializes apps as an empty list", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			cfg := fixture.ReadConfig()
			if apps, ok := cfg["apps"]; ok && apps != nil {
				Expect(apps).To(BeEmpty())
			}
		})
	})

	Describe("idempotency", func() {
		It("succeeds when run a second time", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(fixture.RunSetup()).To(Succeed())
		})

		It("preserves config content on second run", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			firstContent, _ := os.ReadFile(fixture.ConfigPath)

			Expect(fixture.RunSetup()).To(Succeed())
			secondContent, _ := os.ReadFile(fixture.ConfigPath)

			Expect(string(secondContent)).To(Equal(string(firstContent)))
		})
	})

	Describe("--config flag override", func() {
		It("writes config to the path specified by --config", func() {
			customPath := filepath.Join(fixture.TmpDir, "custom", "my-config.yaml")
			Expect(fixture.RunSetup("--config", customPath)).To(Succeed())
			Expect(customPath).To(BeAnExistingFile())
		})
	})

	Describe("stdout output", func() {
		It("prints the config path", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(fixture.Stdout.String()).To(ContainSubstring("Config written to"))
		})

		It("prints the detected architecture", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(fixture.Stdout.String()).To(ContainSubstring(runtime.GOARCH))
		})
	})
})
