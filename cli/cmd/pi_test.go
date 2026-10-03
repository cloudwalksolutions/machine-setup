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

type spyPiAsker struct {
	answer config.PiConfig
	err    error
	inputs forms.PiInputs
}

func (s *spyPiAsker) AskPi(in forms.PiInputs) (config.PiConfig, error) {
	s.inputs = in
	return s.answer, s.err
}

type spyPiOps struct {
	log *[]string
	err error
}

func (o *spyPiOps) Pull(cfg config.PiConfig) error {
	*o.log = append(*o.log, "pull:"+cfg.DefaultModel)
	return o.err
}
func (o *spyPiOps) InstallPackages(p []string) error {
	*o.log = append(*o.log, "install:"+p[0])
	return nil
}

var _ = Describe("PiInit.Run", func() {
	var (
		asker  *spyPiAsker
		store  *memConfigStore
		log    []string
		stdout *bytes.Buffer
		init   *cmd.PiInit
	)

	BeforeEach(func() {
		asker = &spyPiAsker{answer: config.PiConfig{Packages: []string{"npm:pi-subagents"}, OllamaModels: []string{"m"}, DefaultModel: "m"}}
		store = newMemConfigStore("/cfg/config.yaml")
		log = nil
		stdout = &bytes.Buffer{}
		init = &cmd.PiInit{
			Asker: asker,
			Inputs: func() forms.PiInputs {
				return forms.PiInputs{OllamaModels: []string{"m"}, Packages: []string{"npm:pi-subagents"}}
			},
			Config: store,
			Ops:    &spyPiOps{log: &log},
			Report: report.Text{Stdout: stdout, Stderr: &bytes.Buffer{}},
		}
	})

	It("asks with what the machine reports and saves the answers into the config", func() {
		Expect(init.Run()).To(Succeed())

		Expect(asker.inputs).To(Equal(forms.PiInputs{OllamaModels: []string{"m"}, Packages: []string{"npm:pi-subagents"}}))
		Expect(store.cfg.Pi).To(Equal(asker.answer))
		Expect(stdout.String()).To(ContainSubstring("Choices saved to /cfg/config.yaml"))
	})

	It("announces the package install as a step", func() {
		spy := &spyReporter{Text: report.Text{Stdout: stdout, Stderr: &bytes.Buffer{}}}
		init.Report = spy

		Expect(init.Run()).To(Succeed())

		Expect(spy.steps).To(Equal([]string{"Installing pi packages"}))
	})

	It("then pulls and installs the packages, in that order", func() {
		Expect(init.Run()).To(Succeed())

		Expect(log).To(Equal([]string{"pull:m", "install:npm:pi-subagents"}))
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

		p, err := cmd.NewPiInit(report.Text{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}, forms.Headless{})

		Expect(err).NotTo(HaveOccurred())
		Expect(p.Asker).NotTo(BeNil())
		Expect(p.Config).NotTo(BeNil())
		Expect(p.Ops).NotTo(BeNil())
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
		fakeTool("pi", "User packages:\n  npm:pi-llama-cpp\n")
		fakeTool("ollama", "NAME ID SIZE MODIFIED\nqwen2.5-coder:7b x 1 GB now\n")
		seed(".pi/agent/settings.json", `{"theme":"dark","packages":["npm:pi-llama-cpp"]}`)
		seed(".pi/agent/models.json", `{"providers":{"remote":{"baseUrl":"https://llm/v1","api":"openai-completions","apiKey":"$REMOTE_API_KEY"}}}`)
	})

	run := func(args ...string) error {
		orig := os.Args
		DeferCleanup(func() { os.Args = orig })
		os.Args = append([]string{"tars"}, args...)
		return cmd.Execute()
	}

	It("provisions ~/.pi/agent, exposes the local ollama models, keeps other providers, installs only missing packages", func() {
		Expect(run("init", "pi")).To(Succeed())

		Expect(filepath.Join(agent, "agents", "tars.md")).To(BeARegularFile())
		Expect(filepath.Join(agent, "prompts", "tdd.md")).To(BeARegularFile())
		Expect(filepath.Join(agent, "extensions", "block-unreviewable-edits.ts")).To(BeARegularFile())
		Expect(read(filepath.Join(agent, "AGENTS.md"))).To(ContainSubstring("## TDD, strictly"))
		models := read(filepath.Join(agent, "models.json"))
		Expect(models).To(ContainSubstring(`"id": "qwen2.5-coder:7b"`))
		Expect(models).To(ContainSubstring(`"apiKey": "$REMOTE_API_KEY"`))
		Expect(read(filepath.Join(home, "config.yaml"))).To(ContainSubstring("- qwen2.5-coder:7b"))
		Expect(read(filepath.Join(agent, "settings.json"))).To(ContainSubstring(`"defaultThinkingLevel": "off"`))

		calls := read(filepath.Join(home, "fake.log"))
		Expect(calls).To(ContainSubstring("pi install npm:pi-subagents"))
		Expect(calls).To(ContainSubstring("pi install npm:@dietrichgebert/ponytail"))
		Expect(calls).NotTo(ContainSubstring("pi install npm:pi-llama-cpp"))
		Expect(calls).NotTo(ContainSubstring("pi remove"))
	})
})
