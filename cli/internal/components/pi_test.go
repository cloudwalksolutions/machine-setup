package components_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
	"tars/internal/config"
)

var _ = Describe("Pi.Pull", func() {
	var (
		tmp      string
		repoRoot string
		home     string
		agent    string
		opts     components.Options
	)

	write := func(rel, content string) {
		p := filepath.Join(repoRoot, rel)
		Expect(os.MkdirAll(filepath.Dir(p), 0o755)).To(Succeed())
		Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())
	}
	read := func(path string) string {
		b, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred(), path)
		return string(b)
	}

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		repoRoot = filepath.Join(tmp, "repo")
		home = filepath.Join(tmp, "home")
		agent = filepath.Join(home, ".pi", "agent")
		write("pi/settings.json", `{"defaultThinkingLevel":"high","packages":["npm:pi-subagents","npm:bigpowers"]}`)
		write("pi/agents/tars.md", "---\nname: tars\n---\nAGENT")
		write("pi/prompts/tdd.md", "TDD")
		write("pi/prompts/pr.md", "PR")
		write("pi/extensions/block-unreviewable-edits.ts", "EXT")
		write("pi/permissions.json", `{"permission":{"*":"allow"}}`)
		write("pi/keybindings.json", `{"app.clipboard.pasteImage":[]}`)
		write("claude/rules/10-tdd.md", "## TDD\nred first\n")
		write("claude/rules/60-simplicity.md", "## Simple\nyagni\n")

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("copies the agent definition into ~/.pi/agent/agents", func() {
		Expect(components.NewPi(opts).Pull()).To(Succeed())

		Expect(read(filepath.Join(agent, "agents", "tars.md"))).To(Equal("---\nname: tars\n---\nAGENT"))
	})

	It("copies the prompts, the edit-guard extension and the permission baseline", func() {
		Expect(components.NewPi(opts).Pull()).To(Succeed())

		Expect(read(filepath.Join(agent, "prompts", "tdd.md"))).To(Equal("TDD"))
		Expect(read(filepath.Join(agent, "prompts", "pr.md"))).To(Equal("PR"))
		Expect(read(filepath.Join(agent, "extensions", "block-unreviewable-edits.ts"))).To(Equal("EXT"))
		Expect(read(filepath.Join(agent, "extensions", "pi-permission-system", "config.json"))).To(Equal(`{"permission":{"*":"allow"}}`))
	})

	It("merges the keybindings fragment, keeping the user's own bindings", func() {
		Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(agent, "keybindings.json"), []byte(`{"app.session.new":"ctrl+shift+n","app.clipboard.pasteImage":"ctrl+v"}`), 0o644)).To(Succeed())

		Expect(components.NewPi(opts).Pull()).To(Succeed())

		var got map[string]any
		Expect(json.Unmarshal([]byte(read(filepath.Join(agent, "keybindings.json"))), &got)).To(Succeed())
		Expect(got["app.session.new"]).To(Equal("ctrl+shift+n"))
		Expect(got["app.clipboard.pasteImage"]).To(Equal([]any{}))
	})

	It("renders AGENTS.md from the shared rules, honoring the claude rule selection", func() {
		opts.Claude.Rules = []string{"60-simplicity"}
		Expect(components.NewPi(opts).Pull()).To(Succeed())

		got := read(filepath.Join(agent, "AGENTS.md"))
		Expect(got).To(HavePrefix("# Working rules"))
		Expect(got).To(ContainSubstring("## Simple\nyagni\n"))
		Expect(got).NotTo(ContainSubstring("## TDD"))
	})

	It("merges the settings fragment: packages unioned fragment-first, local-only keys kept", func() {
		local := filepath.Join(agent, "settings.json")
		Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
		Expect(os.WriteFile(local, []byte(`{"theme":"dark","packages":["npm:pi-llama-cpp","npm:bigpowers"]}`), 0o644)).To(Succeed())

		Expect(components.NewPi(opts).Pull()).To(Succeed())

		var got map[string]any
		Expect(json.Unmarshal([]byte(read(local)), &got)).To(Succeed())
		Expect(got["theme"]).To(Equal("dark"))
		Expect(got["defaultThinkingLevel"]).To(Equal("high"))
		Expect(got["packages"]).To(Equal([]any{"npm:pi-subagents", "npm:bigpowers", "npm:pi-llama-cpp"}))
	})

	It("renders the ollama provider from config into models.json and pins the default model, keeping other providers", func() {
		Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(agent, "models.json"), []byte(`{"providers":{"other":{"baseUrl":"http://x/v1"}}}`), 0o644)).To(Succeed())
		opts.Pi = config.PiConfig{OllamaModels: []string{"qwen2.5-coder:7b", "llama3.1:8b"}, DefaultModel: "qwen2.5-coder:7b"}

		Expect(components.NewPi(opts).Pull()).To(Succeed())

		var models map[string]any
		Expect(json.Unmarshal([]byte(read(filepath.Join(agent, "models.json"))), &models)).To(Succeed())
		providers := models["providers"].(map[string]any)
		Expect(providers["ollama"]).To(Equal(map[string]any{
			"baseUrl": "http://localhost:11434/v1", "api": "openai-completions", "apiKey": "ollama",
			"models": []any{map[string]any{"id": "qwen2.5-coder:7b"}, map[string]any{"id": "llama3.1:8b"}},
		}))
		Expect(providers).To(HaveKey("other"))

		var settings map[string]any
		Expect(json.Unmarshal([]byte(read(filepath.Join(agent, "settings.json"))), &settings)).To(Succeed())
		Expect(settings["defaultProvider"]).To(Equal("ollama"))
		Expect(settings["defaultModel"]).To(Equal("qwen2.5-coder:7b"))
	})

	It("writes no models.json and pins nothing when no ollama models are configured", func() {
		Expect(components.NewPi(opts).Pull()).To(Succeed())

		Expect(filepath.Join(agent, "models.json")).NotTo(BeAnExistingFile())
		Expect(read(filepath.Join(agent, "settings.json"))).NotTo(ContainSubstring("defaultProvider"))
	})

	It("is a no-op on re-pull: one settings backup, nothing new the second time", func() {
		Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(agent, "settings.json"), []byte(`{"theme":"dark"}`), 0o644)).To(Succeed())
		opts.Pi = config.PiConfig{OllamaModels: []string{"m"}}

		Expect(components.NewPi(opts).Pull()).To(Succeed())
		Expect(components.NewPi(opts).Pull()).To(Succeed())

		Expect(filepath.Join(opts.BackupRoot, "pi", "v1", "settings.json")).To(BeARegularFile())
		Expect(filepath.Join(opts.BackupRoot, "pi", "v2")).NotTo(BeADirectory())
	})

	Describe("InstallPackages", func() {
		It("installs only the packages pi list does not already show", func() {
			var calls []string
			pi := components.NewPi(opts)
			pi.Run = func(name string, args ...string) (string, error) {
				calls = append(calls, name+" "+strings.Join(args, " "))
				if args[0] == "list" {
					return "User packages:\n  npm:pi-llama-cpp\n    /x/node_modules/pi-llama-cpp\n  npm:bigpowers\n    /x/node_modules/bigpowers\n", nil
				}
				return "", nil
			}

			Expect(pi.InstallPackages([]string{"npm:pi-subagents", "npm:bigpowers"})).To(Succeed())

			Expect(calls).To(Equal([]string{"pi list", "pi install npm:pi-subagents"}))
		})
	})

	It("Packages lists the repo fragment's packages", func() {
		Expect(components.NewPi(opts).Packages()).To(Equal([]string{"npm:pi-subagents", "npm:bigpowers"}))
	})

	Describe("OllamaModels", func() {
		It("lists the model names from ollama list", func() {
			pi := components.NewPi(opts)
			pi.Run = func(name string, args ...string) (string, error) {
				Expect(name + " " + strings.Join(args, " ")).To(Equal("ollama list"))
				return "NAME                ID              SIZE      MODIFIED\nqwen2.5-coder:7b    abc123          4.7 GB    2 days ago\nllama3.1:8b         def456          4.9 GB    3 weeks ago\n", nil
			}

			Expect(pi.OllamaModels()).To(Equal([]string{"qwen2.5-coder:7b", "llama3.1:8b"}))
		})
	})

	Describe("Push", func() {
		BeforeEach(func() {
			Expect(components.NewPi(opts).Pull()).To(Succeed())
		})

		It("copies the agent, prompts, extension and permissions back to the repo, archiving under pi-repo", func() {
			Expect(os.WriteFile(filepath.Join(agent, "agents", "tars.md"), []byte("NEW AGENT"), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(agent, "prompts", "tdd.md"), []byte("NEW TDD"), 0o644)).To(Succeed())

			Expect(components.NewPi(opts).Push()).To(Succeed())

			Expect(read(filepath.Join(repoRoot, "pi", "agents", "tars.md"))).To(Equal("NEW AGENT"))
			Expect(read(filepath.Join(repoRoot, "pi", "prompts", "tdd.md"))).To(Equal("NEW TDD"))
			Expect(read(filepath.Join(opts.BackupRoot, "pi-repo", "v1", "tars.md"))).To(Equal("---\nname: tars\n---\nAGENT"))
		})
	})
})
