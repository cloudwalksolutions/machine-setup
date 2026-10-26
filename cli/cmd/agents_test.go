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
	"tars/internal/report"
)

type spyAgentsAsker struct {
	offered []string
	answer  config.AgentsConfig
	err     error
}

func (s *spyAgentsAsker) AskAgents(rules []string) (config.AgentsConfig, error) {
	s.offered = rules
	return s.answer, s.err
}

var _ = Describe("AgentsInit.Run", func() {
	var (
		asker  *spyAgentsAsker
		store  *memConfigStore
		pulled []config.AgentsConfig
		stdout *bytes.Buffer
		init   *cmd.AgentsInit
	)

	BeforeEach(func() {
		asker = &spyAgentsAsker{answer: config.AgentsConfig{SkipRules: []string{"10-tdd"}}}
		store = newMemConfigStore("/cfg/config.yaml")
		pulled = nil
		stdout = &bytes.Buffer{}
		init = &cmd.AgentsInit{
			Asker:  asker,
			Config: store,
			Rules:  []string{"10-tdd", "60-simplicity"},
			Pull:   func(a config.AgentsConfig) error { pulled = append(pulled, a); return nil },
			Report: report.Text{Stdout: stdout, Stderr: &bytes.Buffer{}},
		}
	})

	It("offers the available rules and saves the answer into the config", func() {
		Expect(init.Run()).To(Succeed())

		Expect(asker.offered).To(Equal([]string{"10-tdd", "60-simplicity"}))
		Expect(store.cfg.Agents).To(Equal(asker.answer))
		Expect(stdout.String()).To(ContainSubstring("Choices saved to /cfg/config.yaml"))
	})

	It("pulls with the saved answer", func() {
		Expect(init.Run()).To(Succeed())

		Expect(pulled).To(Equal([]config.AgentsConfig{asker.answer}))
	})

	It("saves nothing and pulls nothing when the user aborts the form", func() {
		asker.err = errors.New("user aborted")

		Expect(init.Run()).To(Succeed())

		Expect(store.cfg.Agents).To(Equal(config.AgentsConfig{}))
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

var _ = Describe("tars init agents (headless, temp HOME)", func() {
	var home string

	BeforeEach(func() {
		home = GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		GinkgoT().Setenv("TARS_CONFIG_PATH", filepath.Join(home, "config.yaml"))
		GinkgoT().Setenv("TARS_BACKUP_ROOT", filepath.Join(home, "backups"))
		GinkgoT().Setenv("TARS_NO_FORM", "1")
	})

	It("renders every rule into CLAUDE.md and AGENTS.md and records the choice", func() {
		orig := os.Args
		DeferCleanup(func() { os.Args = orig })
		os.Args = []string{"tars", "init", "agents"}

		Expect(cmd.Execute()).To(Succeed())

		claudeMD, err := os.ReadFile(filepath.Join(home, ".claude", "CLAUDE.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(claudeMD)).To(ContainSubstring("## Pull requests and CI"))
		agentsMD, err := os.ReadFile(filepath.Join(home, ".pi", "agent", "AGENTS.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(agentsMD)).To(ContainSubstring("## Pull requests and CI"))
		cfg, err := os.ReadFile(filepath.Join(home, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(cfg)).To(ContainSubstring("agents:\n    skip_rules: []"))
	})
})

var _ = Describe("agents composition root", func() {
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("TARS_CONFIG_PATH", filepath.Join(GinkgoT().TempDir(), "config.yaml"))
	})

	It("NewAgentsInit wires the form, config store, repo rules and a puller", func() {
		a, err := cmd.NewAgentsInit(report.Text{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}, forms.Headless{})

		Expect(err).NotTo(HaveOccurred())
		Expect(a.Asker).NotTo(BeNil())
		Expect(a.Config).NotTo(BeNil())
		Expect(a.Pull).NotTo(BeNil())
		Expect(a.Rules).To(ContainElements("10-tdd", "90-reviewable-edits"))
	})
})
