package cmd_test

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/cloudwalk/machine-setup/cmd"
	"github.com/cloudwalk/machine-setup/internal/components"
	"github.com/cloudwalk/machine-setup/internal/config"
	"github.com/cloudwalk/machine-setup/internal/pkg"
)

// ── Test doubles ─────────────────────────────────────────────────────────

type spyWelcome struct{ calls int }

func (s *spyWelcome) Show() error { s.calls++; return nil }

type spyPicker struct {
	offered []string
	pick    []string
}

func (s *spyPicker) Pick(offered []string) ([]string, error) {
	s.offered = offered
	if s.pick != nil {
		return s.pick, nil
	}
	return offered, nil
}

type memConfigStore struct {
	path string
	cfg  *config.Config
}

func newMemConfigStore(path string) *memConfigStore {
	return &memConfigStore{path: path}
}
func (s *memConfigStore) Load() (*config.Config, error) {
	if s.cfg == nil {
		s.cfg = &config.Config{Architecture: "test-arch"}
	}
	return s.cfg, nil
}
func (s *memConfigStore) Save(c *config.Config) error { s.cfg = c; return nil }
func (s *memConfigStore) Path() string                { return s.path }

type fixedRegistry struct{ tools []pkg.Installable }

func (r *fixedRegistry) Installables() []pkg.Installable { return r.tools }
func (r *fixedRegistry) Names() []string {
	names := make([]string, len(r.tools))
	for i, t := range r.tools {
		names[i] = t.Name()
	}
	return names
}

type spyInstallable struct {
	name string
	log  *[]string
	err  error
}

func (s *spyInstallable) Name() string { return s.name }
func (s *spyInstallable) Install(_, _ io.Writer) error {
	*s.log = append(*s.log, s.name)
	return s.err
}

type recordingInstaller struct {
	available []pkg.Installable
	selected  []string
	log       *[]string
	errs      map[string]error
	stderr    io.Writer
}

func (r *recordingInstaller) InstallAll(available []pkg.Installable, selected []string) {
	r.available, r.selected = available, selected
	picked := map[string]bool{}
	for _, n := range selected {
		picked[n] = true
	}
	for _, inst := range available {
		if !picked[inst.Name()] {
			continue
		}
		*r.log = append(*r.log, inst.Name())
		if err := r.errs[inst.Name()]; err != nil {
			fmt.Fprintf(r.stderr, "  %s: %v\n", inst.Name(), err)
		}
	}
}

type spyInstaller struct {
	calls int
	err   error
}

func (s *spyInstaller) Install() error { s.calls++; return s.err }

type spyComponent struct {
	name string
	log  *[]string
	err  error
}

func (c *spyComponent) Name() string { return c.name }
func (c *spyComponent) Pull() error {
	*c.log = append(*c.log, c.name)
	return c.err
}

type recordingPuller struct {
	components []components.Component
	stderr     io.Writer
}

func (p *recordingPuller) PullAll() {
	for _, c := range p.components {
		if err := c.Pull(); err != nil {
			fmt.Fprintf(p.stderr, "  %s: %v\n", c.Name(), err)
		}
	}
}

// ── Fixture ──────────────────────────────────────────────────────────────

// fixture assembles a Setup with test doubles. Tests can read spy state
// directly via the exported fields after calling Run().
type fixture struct {
	Welcome   *spyWelcome
	Picker    *spyPicker
	Config    *memConfigStore
	Installer *recordingInstaller
	OhMyZsh   *spyInstaller
	P10k      *spyInstaller
	Puller    *recordingPuller

	InstallableNames []string
	InstallLog       []string
	InstallErrs      map[string]error

	ComponentNames []string
	PullLog        []string
	ComponentErrs  map[string]error

	Stdout *bytes.Buffer
	Stderr *bytes.Buffer

	Setup *cmd.Setup
}

func newFixture() *fixture {
	f := &fixture{
		InstallableNames: []string{
			"neovim", "byobu", "fzf", "ripgrep", "bat", "eza",
			"jq", "gh", "go", "node", "python",
			"yarn", "n",
			"rustup", "ghcup",
		},
		ComponentNames: []string{"vim", "zsh", "byobu", "nvim", "fonts"},
		InstallErrs:    map[string]error{},
		ComponentErrs:  map[string]error{},
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
	}
	f.assemble()
	return f
}

// assemble builds the spy graph + Setup. Call after mutating InstallErrs /
// ComponentErrs to pick up the new error config.
func (f *fixture) assemble() {
	f.Welcome = &spyWelcome{}
	f.Picker = &spyPicker{}
	f.Config = newMemConfigStore(filepath.Join(GinkgoT().TempDir(), "config.yaml"))
	f.OhMyZsh = &spyInstaller{}
	f.P10k = &spyInstaller{}

	f.InstallLog = []string{}
	tools := make([]pkg.Installable, len(f.InstallableNames))
	for i, n := range f.InstallableNames {
		tools[i] = &spyInstallable{name: n, log: &f.InstallLog, err: f.InstallErrs[n]}
	}
	f.Installer = &recordingInstaller{log: &f.InstallLog, errs: f.InstallErrs, stderr: f.Stderr}

	f.PullLog = []string{}
	comps := make([]components.Component, len(f.ComponentNames))
	for i, n := range f.ComponentNames {
		comps[i] = &spyComponent{name: n, log: &f.PullLog, err: f.ComponentErrs[n]}
	}
	f.Puller = &recordingPuller{components: comps, stderr: f.Stderr}

	f.Setup = &cmd.Setup{
		Welcome:   f.Welcome,
		Picker:    f.Picker,
		Config:    f.Config,
		Registry:  &fixedRegistry{tools: tools},
		Installer: f.Installer,
		OhMyZsh:   f.OhMyZsh,
		P10k:      f.P10k,
		Pull:      f.Puller,
		Stdout:    f.Stdout,
		Stderr:    f.Stderr,
	}
}

// ── Specs ────────────────────────────────────────────────────────────────

var _ = Describe("Setup.Run", func() {
	var f *fixture

	BeforeEach(func() { f = newFixture() })

	Describe("welcome", func() {
		It("invokes the Welcomer exactly once", func() {
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.Welcome.calls).To(Equal(1))
		})
	})

	Describe("tool picker", func() {
		It("offers the registry's names to the picker", func() {
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.Picker.offered).To(Equal(f.InstallableNames))
		})
	})

	Describe("package installation", func() {
		It("installs every selected installable", func() {
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.InstallLog).To(ConsistOf(f.InstallableNames))
		})

		It("continues installing remaining tools when one fails", func() {
			f.InstallErrs = map[string]error{"jq": fmt.Errorf("install failed")}
			f.assemble()

			Expect(f.Setup.Run()).To(Succeed())

			Expect(f.InstallLog).To(ConsistOf(f.InstallableNames))
		})

		It("prints an error line for a failed install without aborting", func() {
			f.InstallErrs = map[string]error{"jq": fmt.Errorf("install failed")}
			f.assemble()

			Expect(f.Setup.Run()).To(Succeed())

			Expect(f.Stderr.String()).To(ContainSubstring("jq"))
		})
	})

	Describe("config persistence", func() {
		It("saves the selected packages by name", func() {
			Expect(f.Setup.Run()).To(Succeed())
			cfg, _ := f.Config.Load()
			names := make([]string, len(cfg.Packages))
			for i, p := range cfg.Packages {
				names[i] = p.Name
			}
			Expect(names).To(Equal(f.InstallableNames))
		})

		It("prints the config path", func() {
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.Stdout.String()).To(ContainSubstring("Config written to"))
			Expect(f.Stdout.String()).To(ContainSubstring(f.Config.Path()))
		})

		It("prints the detected architecture", func() {
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.Stdout.String()).To(ContainSubstring("test-arch"))
		})
	})

	Describe("oh-my-zsh install", func() {
		It("calls the oh-my-zsh installer exactly once", func() {
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.OhMyZsh.calls).To(Equal(1))
		})

		It("reports a failed install without aborting", func() {
			f.OhMyZsh.err = fmt.Errorf("git missing")
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.Stderr.String()).To(ContainSubstring("oh-my-zsh"))
			Expect(f.Stderr.String()).To(ContainSubstring("git missing"))
		})
	})

	Describe("powerlevel10k install", func() {
		It("calls the powerlevel10k installer exactly once", func() {
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.P10k.calls).To(Equal(1))
		})

		It("reports a failed install without aborting", func() {
			f.P10k.err = fmt.Errorf("git clone failed")
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.Stderr.String()).To(ContainSubstring("powerlevel10k"))
		})
	})

	Describe("component pull", func() {
		It("pulls each component in the documented order: vim, zsh, byobu, nvim, fonts", func() {
			Expect(f.Setup.Run()).To(Succeed())
			Expect(f.PullLog).To(Equal(f.ComponentNames))
		})

		It("continues pulling remaining components when one fails, and reports the failure", func() {
			f.ComponentErrs = map[string]error{"zsh": fmt.Errorf("disk full")}
			f.assemble()

			Expect(f.Setup.Run()).To(Succeed())

			Expect(f.PullLog).To(Equal(f.ComponentNames))
			Expect(f.Stderr.String()).To(ContainSubstring("zsh"))
			Expect(f.Stderr.String()).To(ContainSubstring("disk full"))
		})
	})

	Describe("yaml config write", func() {
		It("the saved config can be unmarshaled back as YAML packages", func() {
			Expect(f.Setup.Run()).To(Succeed())
			cfg, err := f.Config.Load()
			Expect(err).NotTo(HaveOccurred())

			// Round-trip the saved config through YAML to make sure the shape stays serializable.
			raw, err := yaml.Marshal(cfg)
			Expect(err).NotTo(HaveOccurred())
			var roundTripped config.Config
			Expect(yaml.Unmarshal(raw, &roundTripped)).To(Succeed())
			Expect(roundTripped.Packages).To(HaveLen(len(f.InstallableNames)))
		})
	})

	Describe("post-setup next steps", func() {
		It("prints the rustup, ghcup, and powerlevel10k hints", func() {
			Expect(f.Setup.Run()).To(Succeed())
			out := f.Stdout.String()
			Expect(out).To(ContainSubstring("rustup install stable"))
			Expect(out).To(ContainSubstring("ghcup tui"))
			Expect(out).To(ContainSubstring("Powerlevel10k"))
		})
	})
})
