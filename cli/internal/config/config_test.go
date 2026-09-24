package config_test

import (
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/config"
)

var _ = Describe("DefaultConfigPath", func() {
	It("honors the TARS_CONFIG_PATH override", func() {
		GinkgoT().Setenv("TARS_CONFIG_PATH", "/tmp/custom-config.yaml")
		Expect(config.DefaultConfigPath()).To(Equal("/tmp/custom-config.yaml"))
	})

	It("defaults to ~/.config/tars/config.yaml", func() {
		GinkgoT().Setenv("TARS_CONFIG_PATH", "")
		GinkgoT().Setenv("HOME", "/fake/home")
		Expect(config.DefaultConfigPath()).To(Equal("/fake/home/.config/tars/config.yaml"))
	})
})

var _ = Describe("Init", func() {
	var path string

	BeforeEach(func() {
		path = filepath.Join(GinkgoT().TempDir(), "nested", "config.yaml")
	})

	It("creates the file with defaults when absent, creating parent dirs", func() {
		cfg, err := config.Init(path)

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Architecture).To(Equal(runtime.GOARCH))
		Expect(cfg.Packages).To(BeEmpty())
		Expect(path).To(BeARegularFile())
	})

	It("loads an existing config, preserving user-set values", func() {
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, []byte("architecture: riscv\npackages:\n  - name: jq\n    manager: brew\n"), 0o644)).To(Succeed())

		cfg, err := config.Init(path)

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Architecture).To(Equal("riscv"))
		Expect(cfg.Packages).To(HaveLen(1))
		Expect(cfg.Packages[0].Name).To(Equal("jq"))
	})

	It("is idempotent — a second Init returns the same config", func() {
		first, err := config.Init(path)
		Expect(err).NotTo(HaveOccurred())

		second, err := config.Init(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(second).To(Equal(first))
	})

	It("returns an error for an unparseable file", func() {
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, []byte(":\tnot yaml {{{"), 0o644)).To(Succeed())

		_, err := config.Init(path)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Save", func() {
	var path string

	BeforeEach(func() {
		path = filepath.Join(GinkgoT().TempDir(), "config.yaml")
	})

	It("round-trips a config through Save and Init", func() {
		cfg := &config.Config{
			Architecture: "arm64",
			Packages:     []config.Package{{Name: "jq", Manager: "brew"}},
		}

		Expect(config.Save(path, cfg)).To(Succeed())

		loaded, err := config.Init(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(loaded.Architecture).To(Equal("arm64"))
		Expect(loaded.Packages).To(Equal(cfg.Packages))
	})

	It("preserves unrecognized keys already in the file", func() {
		Expect(os.WriteFile(path, []byte("architecture: arm64\ncustom_key: keep-me\n"), 0o644)).To(Succeed())

		Expect(config.Save(path, &config.Config{Architecture: "amd64"})).To(Succeed())

		b, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(ContainSubstring("keep-me"))
		Expect(string(b)).To(ContainSubstring("amd64"))
	})

	It("round-trips the pi section", func() {
		keep := false
		cfg := &config.Config{Pi: config.PiConfig{
			Packages:     []string{"npm:pi-subagents"},
			RemoveGentle: &keep,
			Providers: []config.PiProvider{
				{Name: "ollama", Kind: "ollama", Models: []string{"qwen2.5-coder:7b"}},
				{Name: "remote", Kind: "openai", BaseURL: "https://llm.example/v1", KeyEnv: "REMOTE_LLM_API_KEY", Models: []string{"qwen3-14b"}},
			},
			DefaultProvider: "remote",
			DefaultModel:    "qwen3-14b",
		}}

		Expect(config.Save(path, cfg)).To(Succeed())

		loaded, err := config.Init(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(loaded.Pi).To(Equal(cfg.Pi))
	})

	It("round-trips the claude section", func() {
		off := false
		cfg := &config.Config{Claude: config.ClaudeConfig{Hook: &off, Rules: []string{"10-tdd", "60-simplicity"}}}

		Expect(config.Save(path, cfg)).To(Succeed())

		loaded, err := config.Init(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(*loaded.Claude.Hook).To(BeFalse())
		Expect(loaded.Claude.Settings).To(BeNil())
		Expect(loaded.Claude.Rules).To(Equal([]string{"10-tdd", "60-simplicity"}))
	})
})
