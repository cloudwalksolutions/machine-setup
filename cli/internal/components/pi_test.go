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

	It("renders configured providers into models.json and sets the defaults, keeping local-only providers", func() {
		Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(agent, "models.json"), []byte(`{"providers":{"remote":{"baseUrl":"https://old/v1","api":"openai-completions","apiKey":"literal-secret","models":[{"id":"old"}]},"other":{"baseUrl":"http://x/v1"}}}`), 0o644)).To(Succeed())
		opts.Pi = config.PiConfig{
			Providers: []config.PiProvider{
				{Name: "ollama", Kind: "ollama", Models: []string{"qwen2.5-coder:7b"}},
				{Name: "remote", Kind: "openai", BaseURL: "https://llm.example/v1", KeyEnv: "REMOTE_LLM_API_KEY", Models: []string{"qwen3-14b"}},
			},
			DefaultProvider: "remote",
			DefaultModel:    "qwen3-14b",
		}

		Expect(components.NewPi(opts).Pull()).To(Succeed())

		var models map[string]any
		Expect(json.Unmarshal([]byte(read(filepath.Join(agent, "models.json"))), &models)).To(Succeed())
		providers := models["providers"].(map[string]any)
		Expect(providers["ollama"]).To(Equal(map[string]any{
			"baseUrl": "http://localhost:11434/v1", "api": "openai-completions", "apiKey": "ollama",
			"models": []any{map[string]any{"id": "qwen2.5-coder:7b"}},
		}))
		Expect(providers["remote"]).To(Equal(map[string]any{
			"baseUrl": "https://llm.example/v1", "api": "openai-completions", "apiKey": "$REMOTE_LLM_API_KEY",
			"models": []any{map[string]any{"id": "qwen3-14b"}},
		}))
		Expect(providers).To(HaveKey("other"))

		var settings map[string]any
		Expect(json.Unmarshal([]byte(read(filepath.Join(agent, "settings.json"))), &settings)).To(Succeed())
		Expect(settings["defaultProvider"]).To(Equal("remote"))
		Expect(settings["defaultModel"]).To(Equal("qwen3-14b"))
	})

	It("points pi-llama-cpp at the server via llamaServerUrl instead of a models.json provider", func() {
		opts.Pi = config.PiConfig{Providers: []config.PiProvider{{Name: "local", Kind: "llamacpp", BaseURL: "http://localhost:8080"}}}

		Expect(components.NewPi(opts).Pull()).To(Succeed())

		var settings map[string]any
		Expect(json.Unmarshal([]byte(read(filepath.Join(agent, "settings.json"))), &settings)).To(Succeed())
		Expect(settings["llamaServerUrl"]).To(Equal("http://localhost:8080"))
		Expect(filepath.Join(agent, "models.json")).NotTo(BeAnExistingFile())
	})

	It("is a no-op on re-pull: one settings backup, nothing new the second time", func() {
		Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(agent, "settings.json"), []byte(`{"theme":"dark"}`), 0o644)).To(Succeed())
		opts.Pi = config.PiConfig{Providers: []config.PiProvider{{Name: "ollama", Kind: "ollama", Models: []string{"m"}}}}

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

	Describe("DetectProviders", func() {
		It("reads existing models.json providers as openai entries with a derived key env name", func() {
			Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(agent, "models.json"), []byte(`{"providers":{"remote-llama":{"baseUrl":"https://llm/v1","api":"openai-completions","apiKey":"lit","models":[{"id":"a"},{"id":"b"}]},"ollama":{"baseUrl":"http://localhost:11434/v1","apiKey":"ollama","models":[{"id":"q"}]}}}`), 0o644)).To(Succeed())

			got := components.NewPi(opts).DetectProviders()

			Expect(got).To(ConsistOf(
				config.PiProvider{Name: "remote-llama", Kind: "openai", BaseURL: "https://llm/v1", KeyEnv: "REMOTE_LLAMA_API_KEY", Models: []string{"a", "b"}},
				config.PiProvider{Name: "ollama", Kind: "ollama", BaseURL: "http://localhost:11434/v1", Models: []string{"q"}},
			))
		})

		It("is empty without a models.json", func() {
			Expect(components.NewPi(opts).DetectProviders()).To(BeEmpty())
		})
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

	Describe("RemoveGentle", func() {
		It("removes the package and backs up then deletes the files gentle-pi seeded, keeping ours", func() {
			for _, rel := range []string{
				".pi/agent/agents/sdd-init.md", ".pi/agent/agents/jd-judge-a.md", ".pi/agent/agents/review-risk.md",
				".pi/agent/agents/gentle-ai-worker.md", ".pi/agent/agents/tars.md",
				".pi/agent/chains/sdd-full.chain.md", ".pi/agent/gentle-ai/managed-assets.json",
				".pi/agent/gentle-agents/presence/x", ".pi/gentle-ai/persona.json",
			} {
				p := filepath.Join(home, rel)
				Expect(os.MkdirAll(filepath.Dir(p), 0o755)).To(Succeed())
				Expect(os.WriteFile(p, []byte(rel), 0o644)).To(Succeed())
			}
			var calls []string
			pi := components.NewPi(opts)
			pi.Run = func(name string, args ...string) (string, error) {
				calls = append(calls, name+" "+strings.Join(args, " "))
				return "", nil
			}

			Expect(pi.RemoveGentle()).To(Succeed())

			Expect(calls).To(Equal([]string{"pi remove npm:gentle-pi"}))
			for _, gone := range []string{"agents/sdd-init.md", "agents/jd-judge-a.md", "agents/review-risk.md", "agents/gentle-ai-worker.md", "chains", "gentle-ai", "gentle-agents"} {
				Expect(filepath.Join(agent, gone)).NotTo(BeAnExistingFile(), gone)
			}
			Expect(filepath.Join(home, ".pi", "gentle-ai")).NotTo(BeAnExistingFile())
			Expect(filepath.Join(agent, "agents", "tars.md")).To(BeARegularFile())
			backedUp := func(rel string) []string {
				m, err := filepath.Glob(filepath.Join(opts.BackupRoot, "pi", "v*", rel))
				Expect(err).NotTo(HaveOccurred())
				return m
			}
			Expect(backedUp("sdd-init.md")).To(HaveLen(1))
			Expect(read(backedUp("sdd-init.md")[0])).To(Equal(".pi/agent/agents/sdd-init.md"))
			Expect(backedUp(filepath.Join("chains", "sdd-full.chain.md"))).To(HaveLen(1))
		})
	})

	Describe("RelocateKeys", func() {
		It("moves a literal apiKey into ~/.zshrc_secret as the provider's env var", func() {
			Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(agent, "models.json"), []byte(`{"providers":{"remote":{"apiKey":"literal-secret"}}}`), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(home, ".zshrc_secret"), []byte("# secrets\n"), 0o644)).To(Succeed())
			opts.Pi = config.PiConfig{Providers: []config.PiProvider{{Name: "remote", Kind: "openai", BaseURL: "https://llm/v1", KeyEnv: "REMOTE_LLM_API_KEY"}}}

			Expect(components.NewPi(opts).RelocateKeys()).To(Succeed())

			Expect(read(filepath.Join(home, ".zshrc_secret"))).To(Equal("# secrets\nexport REMOTE_LLM_API_KEY=\"literal-secret\"\n"))
			Expect(read(filepath.Join(opts.BackupRoot, "pi", "v1", ".zshrc_secret"))).To(Equal("# secrets\n"))
		})

		It("leaves the secret file alone when the key is already an env reference", func() {
			Expect(os.MkdirAll(agent, 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(agent, "models.json"), []byte(`{"providers":{"remote":{"apiKey":"$REMOTE_LLM_API_KEY"}}}`), 0o644)).To(Succeed())
			opts.Pi = config.PiConfig{Providers: []config.PiProvider{{Name: "remote", Kind: "openai", KeyEnv: "REMOTE_LLM_API_KEY"}}}

			Expect(components.NewPi(opts).RelocateKeys()).To(Succeed())

			Expect(filepath.Join(home, ".zshrc_secret")).NotTo(BeAnExistingFile())
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
