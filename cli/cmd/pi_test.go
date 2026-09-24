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
)

type spyPiAsker struct {
	answer config.PiConfig
	err    error
}

func (s *spyPiAsker) Ask() (config.PiConfig, error) { return s.answer, s.err }

type spyPiOps struct {
	log *[]string
	err error
}

func (o *spyPiOps) Pull(cfg config.PiConfig) error {
	*o.log = append(*o.log, "pull:"+cfg.DefaultModel)
	return o.err
}
func (o *spyPiOps) RelocateKeys(config.PiConfig) error {
	*o.log = append(*o.log, "relocate")
	return nil
}
func (o *spyPiOps) InstallPackages(p []string) error {
	*o.log = append(*o.log, "install:"+p[0])
	return nil
}
func (o *spyPiOps) RemoveGentle() error { *o.log = append(*o.log, "remove-gentle"); return nil }

var _ = Describe("PiInit.Run", func() {
	var (
		asker *spyPiAsker
		store *memConfigStore
		log   []string
		init  *cmd.PiInit
	)

	BeforeEach(func() {
		yes := true
		asker = &spyPiAsker{answer: config.PiConfig{Packages: []string{"npm:pi-subagents"}, RemoveGentle: &yes, DefaultModel: "m"}}
		store = newMemConfigStore("/cfg/config.yaml")
		log = nil
		init = &cmd.PiInit{Asker: asker, Config: store, Ops: &spyPiOps{log: &log}, Stdout: &bytes.Buffer{}}
	})

	It("saves the answers into the config", func() {
		Expect(init.Run()).To(Succeed())

		Expect(store.cfg.Pi).To(Equal(asker.answer))
	})

	It("then relocates keys before pulling, installs the packages and removes gentle-pi, in that order", func() {
		Expect(init.Run()).To(Succeed())

		Expect(log).To(Equal([]string{"relocate", "pull:m", "install:npm:pi-subagents", "remove-gentle"}))
	})

	It("saves nothing and runs nothing when the user aborts the form", func() {
		asker.err = errors.New("user aborted")

		Expect(init.Run()).To(Succeed())

		Expect(store.cfg.Pi).To(Equal(config.PiConfig{}))
		Expect(log).To(BeEmpty())
	})

	It("NewPiInit wires the form, the config store and the machine ops", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("TARS_CONFIG_PATH", filepath.Join(GinkgoT().TempDir(), "config.yaml"))

		p, err := cmd.NewPiInit(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(p.Asker).NotTo(BeNil())
		Expect(p.Config).NotTo(BeNil())
		Expect(p.Ops).NotTo(BeNil())
	})

	It("keeps gentle-pi when the form said so", func() {
		no := false
		asker.answer.RemoveGentle = &no

		Expect(init.Run()).To(Succeed())

		Expect(log).NotTo(ContainElement("remove-gentle"))
	})
})

var _ = Describe("tars pi init (headless, temp HOME, fake pi and ollama on PATH)", func() {
	var (
		home  string
		agent string
		bin   string
	)

	fakeTool := func(name, stdout string) {
		script := "#!/bin/sh\necho \"" + name + " $*\" >> \"$FAKE_LOG\"\nprintf '%s' '" + stdout + "'\n"
		Expect(os.WriteFile(filepath.Join(bin, name), []byte(script), 0o755)).To(Succeed())
	}
	seed := func(rel, content string) {
		p := filepath.Join(home, rel)
		Expect(os.MkdirAll(filepath.Dir(p), 0o755)).To(Succeed())
		Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())
	}
	read := func(path string) string {
		b, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred(), path)
		return string(b)
	}

	BeforeEach(func() {
		home = GinkgoT().TempDir()
		agent = filepath.Join(home, ".pi", "agent")
		bin = filepath.Join(home, "bin")
		Expect(os.MkdirAll(bin, 0o755)).To(Succeed())
		GinkgoT().Setenv("HOME", home)
		GinkgoT().Setenv("TARS_CONFIG_PATH", filepath.Join(home, "config.yaml"))
		GinkgoT().Setenv("TARS_BACKUP_ROOT", filepath.Join(home, "backups"))
		GinkgoT().Setenv("TARS_NO_FORM", "1")
		GinkgoT().Setenv("FAKE_LOG", filepath.Join(home, "fake.log"))
		GinkgoT().Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
		fakeTool("pi", "User packages:\n  npm:pi-llama-cpp\n  npm:gentle-pi\n")
		fakeTool("ollama", "NAME ID SIZE MODIFIED\nqwen2.5-coder:7b x 1 GB now\n")
		seed(".pi/agent/settings.json", `{"theme":"dark","packages":["npm:pi-llama-cpp","npm:gentle-pi"]}`)
		seed(".pi/agent/models.json", `{"providers":{"remote-llama":{"baseUrl":"https://llm/v1","api":"openai-completions","apiKey":"literal-secret","models":[{"id":"qwen3-14b"}]}}}`)
		seed(".pi/agent/agents/sdd-init.md", "gentle")
		seed(".zshrc_secret", "# secrets\n")
	})

	run := func(args ...string) error {
		orig := os.Args
		DeferCleanup(func() { os.Args = orig })
		os.Args = append([]string{"tars"}, args...)
		return cmd.Execute()
	}

	It("provisions ~/.pi/agent, relocates the key, installs the missing packages and removes gentle-pi", func() {
		Expect(run("pi", "init")).To(Succeed())

		Expect(filepath.Join(agent, "agents", "tars.md")).To(BeARegularFile())
		Expect(filepath.Join(agent, "prompts", "tdd.md")).To(BeARegularFile())
		Expect(filepath.Join(agent, "extensions", "block-unreviewable-edits.ts")).To(BeARegularFile())
		Expect(read(filepath.Join(agent, "AGENTS.md"))).To(ContainSubstring("## TDD, strictly"))
		Expect(read(filepath.Join(agent, "models.json"))).To(ContainSubstring(`"apiKey": "$REMOTE_LLAMA_API_KEY"`))
		Expect(read(filepath.Join(agent, "models.json"))).NotTo(ContainSubstring("literal-secret"))
		Expect(read(filepath.Join(home, ".zshrc_secret"))).To(ContainSubstring(`export REMOTE_LLAMA_API_KEY="literal-secret"`))
		Expect(read(filepath.Join(home, "config.yaml"))).To(ContainSubstring("key_env: REMOTE_LLAMA_API_KEY"))
		Expect(filepath.Join(agent, "agents", "sdd-init.md")).NotTo(BeAnExistingFile())

		calls := read(filepath.Join(home, "fake.log"))
		Expect(calls).To(ContainSubstring("pi install npm:pi-subagents"))
		Expect(calls).To(ContainSubstring("pi install npm:@gotgenes/pi-permission-system"))
		Expect(calls).NotTo(ContainSubstring("pi install npm:pi-llama-cpp"))
		Expect(calls).To(ContainSubstring("pi remove npm:gentle-pi"))
	})
})
