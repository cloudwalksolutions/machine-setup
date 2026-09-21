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
	It("honors the MACHINE_SETUP_CONFIG_PATH override", func() {
		GinkgoT().Setenv("MACHINE_SETUP_CONFIG_PATH", "/tmp/custom-config.yaml")
		Expect(config.DefaultConfigPath()).To(Equal("/tmp/custom-config.yaml"))
	})

	It("defaults to ~/.config/.machine-setup/config.yaml", func() {
		GinkgoT().Setenv("MACHINE_SETUP_CONFIG_PATH", "")
		GinkgoT().Setenv("HOME", "/fake/home")
		Expect(config.DefaultConfigPath()).To(Equal("/fake/home/.config/.machine-setup/config.yaml"))
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
})
