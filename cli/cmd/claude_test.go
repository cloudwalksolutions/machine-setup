package cmd_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/config"
	"tars/internal/forms"
)

type spyClaudeAsker struct {
	offered []string
	answer  config.ClaudeConfig
	err     error
}

func (s *spyClaudeAsker) Ask(rules []string) (config.ClaudeConfig, error) {
	s.offered = rules
	return s.answer, s.err
}

var _ = Describe("ClaudeInit.Run", func() {
	var (
		asker  *spyClaudeAsker
		store  *memConfigStore
		pulled []config.ClaudeConfig
		init   *cmd.ClaudeInit
	)

	BeforeEach(func() {
		off := false
		asker = &spyClaudeAsker{answer: config.ClaudeConfig{Hook: &off, Rules: []string{"10-tdd"}}}
		store = newMemConfigStore("/cfg/config.yaml")
		pulled = nil
		init = &cmd.ClaudeInit{
			Asker:  asker,
			Config: store,
			Rules:  []string{"10-tdd", "60-simplicity"},
			Pull:   func(c config.ClaudeConfig) error { pulled = append(pulled, c); return nil },
			Stdout: &bytes.Buffer{},
		}
	})

	It("offers the available rules and saves the answers into the config", func() {
		Expect(init.Run()).To(Succeed())

		Expect(asker.offered).To(Equal([]string{"10-tdd", "60-simplicity"}))
		Expect(store.cfg.Claude).To(Equal(asker.answer))
	})

	It("pulls with the saved answers", func() {
		Expect(init.Run()).To(Succeed())

		Expect(pulled).To(Equal([]config.ClaudeConfig{asker.answer}))
	})

	It("saves nothing and pulls nothing when the user aborts the form", func() {
		asker.err = errors.New("user aborted")

		Expect(init.Run()).To(Succeed())

		Expect(store.cfg.Claude).To(Equal(config.ClaudeConfig{}))
		Expect(pulled).To(BeEmpty())
	})

	It("surfaces form failures other than an abort", func() {
		asker.err = errors.New("tty gone")

		Expect(init.Run()).To(MatchError(ContainSubstring("tty gone")))
		Expect(pulled).To(BeEmpty())
	})

	It("surfaces a config save failure before pulling", func() {
		init.Config = &failingConfigStore{saveErr: errors.New("disk full")}

		Expect(init.Run()).To(MatchError(ContainSubstring("disk full")))
		Expect(pulled).To(BeEmpty())
	})
})

type failingConfigStore struct{ saveErr error }

func (f *failingConfigStore) Load() (*config.Config, error) { return &config.Config{}, nil }
func (f *failingConfigStore) Save(*config.Config) error     { return f.saveErr }
func (f *failingConfigStore) Path() string                  { return "/cfg/config.yaml" }

type spyProjectAsker struct {
	answer forms.ProjectAnswers
	err    error
}

func (s *spyProjectAsker) Ask(defaultName string) (forms.ProjectAnswers, error) {
	if s.answer.Name == "" {
		s.answer.Name = defaultName
	}
	return s.answer, s.err
}

var _ = Describe("ClaudeProject.Run", func() {
	var (
		tmp     string
		dir     string
		project *cmd.ClaudeProject
	)

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		dir = filepath.Join(tmp, "my-app")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		tpl := filepath.Join(tmp, "CLAUDE.project.md")
		Expect(os.WriteFile(tpl, []byte("# {{.Name}}\n\n{{.Description}}\n\n```\n{{.TestCommand}}\n```\n"), 0o644)).To(Succeed())
		project = &cmd.ClaudeProject{
			Asker:      &spyProjectAsker{answer: forms.ProjectAnswers{Description: "Does things", TestCommand: "make test"}},
			Template:   tpl,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
		}
	})

	It("renders the template into <dir>/CLAUDE.md, defaulting the name to the dir's basename", func() {
		Expect(project.Run(dir)).To(Succeed())

		b, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("# my-app\n\nDoes things\n\n```\nmake test\n```\n"))
	})

	It("refuses to overwrite an existing CLAUDE.md unless forced", func() {
		existing := filepath.Join(dir, "CLAUDE.md")
		Expect(os.WriteFile(existing, []byte("KEEP"), 0o644)).To(Succeed())

		err := project.Run(dir)

		Expect(err).To(MatchError(ContainSubstring("--force")))
		b, _ := os.ReadFile(existing)
		Expect(string(b)).To(Equal("KEEP"))
	})

	It("fails when the template is missing", func() {
		project.Template = filepath.Join(tmp, "nope.md")

		Expect(project.Run(dir)).NotTo(Succeed())
		Expect(filepath.Join(dir, "CLAUDE.md")).NotTo(BeAnExistingFile())
	})

	It("backs up the existing CLAUDE.md under claude-project before overwriting when forced", func() {
		Expect(os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("OLD"), 0o644)).To(Succeed())
		project.Force = true

		Expect(project.Run(dir)).To(Succeed())

		b, err := os.ReadFile(filepath.Join(project.BackupRoot, "claude-project", "v1", "CLAUDE.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("OLD"))
		b, _ = os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
		Expect(string(b)).To(HavePrefix("# my-app"))
	})
})

var _ = Describe("LoadClaudeConfig", func() {
	It("returns the saved claude section from the config file", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		GinkgoT().Setenv("TARS_CONFIG_PATH", path)
		Expect(os.WriteFile(path, []byte("claude:\n  hook: false\n  rules: [10-tdd]\n"), 0o644)).To(Succeed())

		c, err := cmd.LoadClaudeConfig()

		Expect(err).NotTo(HaveOccurred())
		Expect(*c.Hook).To(BeFalse())
		Expect(c.Rules).To(Equal([]string{"10-tdd"}))
	})

	It("returns the all-on zero value when no config exists yet", func() {
		GinkgoT().Setenv("TARS_CONFIG_PATH", filepath.Join(GinkgoT().TempDir(), "config.yaml"))

		c, err := cmd.LoadClaudeConfig()

		Expect(err).NotTo(HaveOccurred())
		Expect(c).To(Equal(config.ClaudeConfig{}))
	})
})

var _ = Describe("tars claude (headless, temp HOME)", func() {
	var home string

	BeforeEach(func() {
		home = GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		GinkgoT().Setenv("TARS_CONFIG_PATH", filepath.Join(home, "config.yaml"))
		GinkgoT().Setenv("TARS_BACKUP_ROOT", filepath.Join(home, "backups"))
		GinkgoT().Setenv("TARS_NO_FORM", "1")
	})

	run := func(args ...string) error {
		orig := os.Args
		DeferCleanup(func() { os.Args = orig })
		os.Args = append([]string{"tars"}, args...)
		return cmd.Execute()
	}

	It("init provisions ~/.claude with every rule and records the choices", func() {
		Expect(run("claude", "init")).To(Succeed())

		Expect(filepath.Join(home, ".claude", "hooks", "block-unreviewable-edits.sh")).To(BeARegularFile())
		Expect(filepath.Join(home, ".claude", "settings.json")).To(BeARegularFile())
		b, err := os.ReadFile(filepath.Join(home, ".claude", "CLAUDE.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(ContainSubstring("## Pull requests and CI"))
		cfg, err := os.ReadFile(filepath.Join(home, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(cfg)).To(ContainSubstring("20-prs-and-ci"))
	})

	It("project scaffolds CLAUDE.md and refuses a second run without --force", func() {
		dir := filepath.Join(home, "proj")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())

		Expect(run("claude", "project", dir)).To(Succeed())
		b, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(ContainSubstring("working in proj."))

		Expect(run("claude", "project", dir)).To(MatchError(ContainSubstring("--force")))
		Expect(run("claude", "project", "--force", dir)).To(Succeed())
		Expect(filepath.Join(home, "backups", "claude-project", "v1", "CLAUDE.md")).To(BeARegularFile())
	})
})

var _ = Describe("claude composition roots", func() {
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("TARS_CONFIG_PATH", filepath.Join(GinkgoT().TempDir(), "config.yaml"))
	})

	It("NewClaudeInit wires the form, config store, repo rules and a puller", func() {
		c, err := cmd.NewClaudeInit(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(c.Asker).NotTo(BeNil())
		Expect(c.Config).NotTo(BeNil())
		Expect(c.Pull).NotTo(BeNil())
		Expect(c.Rules).To(ContainElements("10-tdd", "90-reviewable-edits"))
	})

	It("NewClaudeProject points at the repo template", func() {
		p, err := cmd.NewClaudeProject(&bytes.Buffer{}, &bytes.Buffer{}, true)

		Expect(err).NotTo(HaveOccurred())
		Expect(p.Asker).NotTo(BeNil())
		Expect(p.Force).To(BeTrue())
		Expect(p.Template).To(HaveSuffix(filepath.Join("claude", "templates", "CLAUDE.project.md")))
	})
})
