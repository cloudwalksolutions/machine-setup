package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/cloudwalk/machine-setup/internal/pkg"
)

// ManagerSpy is a test spy that records all calls and allows configuring responses.
type ManagerSpy struct {
	InstallCalls    []string
	UninstallCalls  []string
	UpdateCalls     []string
	IsInstalledCalls []string

	InstallErrors      map[string]error
	UninstallErrors    map[string]error
	UpdateErrors       map[string]error
	IsInstalledReturns map[string]bool
}

func NewManagerSpy() *ManagerSpy {
	return &ManagerSpy{
		InstallErrors:      make(map[string]error),
		UninstallErrors:    make(map[string]error),
		UpdateErrors:       make(map[string]error),
		IsInstalledReturns: make(map[string]bool),
	}
}

func (s *ManagerSpy) Install(name string) error {
	s.InstallCalls = append(s.InstallCalls, name)
	return s.InstallErrors[name]
}

func (s *ManagerSpy) Uninstall(name string) error {
	s.UninstallCalls = append(s.UninstallCalls, name)
	return s.UninstallErrors[name]
}

func (s *ManagerSpy) Update(name string) error {
	s.UpdateCalls = append(s.UpdateCalls, name)
	return s.UpdateErrors[name]
}

func (s *ManagerSpy) IsInstalled(name string) (bool, error) {
	s.IsInstalledCalls = append(s.IsInstalledCalls, name)
	return s.IsInstalledReturns[name], nil
}

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

	// Always inject a no-op spy so no spec accidentally hits real brew.
	newManagerFn = func(stdout, stderr io.Writer) (pkg.Manager, error) {
		return NewManagerSpy(), nil
	}
}

func (f *ExecutorFixture) Teardown() {
	os.RemoveAll(f.TmpDir)
	os.Unsetenv("MACHINE_SETUP_CONFIG_PATH")
	os.Unsetenv("MACHINE_SETUP_NO_FORM")
	newManagerFn = pkg.NewManager
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

	Describe("package installation", func() {
		var spy *ManagerSpy

		BeforeEach(func() {
			spy = NewManagerSpy()
			newManagerFn = func(stdout, stderr io.Writer) (pkg.Manager, error) {
				return spy, nil
			}
		})

		It("installs all selected dev tools", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(spy.InstallCalls).To(ConsistOf(pkg.DevToolNames()))
		})

		It("prints an installing message for each tool", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			for _, name := range pkg.DevToolNames() {
				Expect(fixture.Stdout.String()).To(ContainSubstring("Installing " + name))
			}
		})

		It("saves the selected packages to the config file", func() {
			Expect(fixture.RunSetup()).To(Succeed())
			cfg := fixture.ReadConfig()
			pkgs, ok := cfg["packages"]
			Expect(ok).To(BeTrue())
			Expect(pkgs).To(HaveLen(len(pkg.DevTools)))
		})

		It("continues installing remaining tools when one fails", func() {
			spy.InstallErrors["jq"] = fmt.Errorf("install failed")
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(spy.InstallCalls).To(ConsistOf(pkg.DevToolNames()))
		})

		It("prints an error line for a failed install without aborting", func() {
			spy.InstallErrors["jq"] = fmt.Errorf("install failed")
			Expect(fixture.RunSetup()).To(Succeed())
			Expect(fixture.Stderr.String()).To(ContainSubstring("jq"))
		})
	})
})
