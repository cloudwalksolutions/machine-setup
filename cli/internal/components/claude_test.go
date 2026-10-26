package components_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
	"tars/internal/config"
)

var _ = Describe("Claude.Pull", func() {
	var (
		tmp      string
		repoRoot string
		home     string
		opts     components.Options
	)

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		repoRoot = filepath.Join(tmp, "repo")
		home = filepath.Join(tmp, "home")
		Expect(os.MkdirAll(filepath.Join(repoRoot, "claude", "hooks"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(repoRoot, "agents", "rules"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "claude", "hooks", "block-unreviewable-edits.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "claude", "settings.json"), []byte(`{"model":"m","hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"$HOME/.claude/hooks/block-unreviewable-edits.sh"}]}]}}`), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "agents", "rules", "10-tdd.md"), []byte("## TDD\nred first\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "agents", "rules", "60-simplicity.md"), []byte("## Simple\nyagni\n"), 0o644)).To(Succeed())

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("writes nothing on a dry run", func() {
		opts.DryRun = true

		Expect(components.NewClaude(opts).Pull()).To(Succeed())

		_, err := os.Stat(home)
		Expect(os.IsNotExist(err)).To(BeTrue(), "dry-run must not create HOME")
	})

	It("shows the diff of a changed rule on a diff run", func() {
		Expect(components.NewClaude(opts).Pull()).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "agents", "rules", "10-tdd.md"), []byte("## TDD\nred first, always\n"), 0o644)).To(Succeed())
		out := &bytes.Buffer{}
		opts.DryRun, opts.Diff, opts.Stdout = true, true, out

		Expect(components.NewClaude(opts).Pull()).To(Succeed())

		Expect(out.String()).To(ContainSubstring("-red first\n+red first, always\n"))
	})

	It("copies the hook into ~/.claude/hooks keeping it executable", func() {
		Expect(components.NewClaude(opts).Pull()).To(Succeed())

		info, err := os.Stat(filepath.Join(home, ".claude", "hooks", "block-unreviewable-edits.sh"))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm() & 0o111).NotTo(BeZero())
	})

	It("renders every rule file, in name order, into ~/.claude/CLAUDE.md by default", func() {
		Expect(components.NewClaude(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".claude", "CLAUDE.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(HavePrefix("# Global Claude Code rules"))
		Expect(string(b)).To(MatchRegexp(`(?s)## TDD\nred first\n.*## Simple\nyagni\n`))
	})

	It("renders every rule except the skipped ones, so new rule files show up without re-init", func() {
		opts.Agents = config.AgentsConfig{SkipRules: []string{"10-tdd"}}
		Expect(components.NewClaude(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".claude", "CLAUDE.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(ContainSubstring("## Simple"))
		Expect(string(b)).NotTo(ContainSubstring("## TDD"))
	})

	It("merges the settings fragment into an existing settings.json, keeping local-only keys", func() {
		local := filepath.Join(home, ".claude", "settings.json")
		Expect(os.MkdirAll(filepath.Dir(local), 0o755)).To(Succeed())
		Expect(os.WriteFile(local, []byte(`{"model":"old","autoMode":{"allow":["x"]}}`), 0o644)).To(Succeed())

		Expect(components.NewClaude(opts).Pull()).To(Succeed())

		var got map[string]any
		b, err := os.ReadFile(local)
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Unmarshal(b, &got)).To(Succeed())
		Expect(got["model"]).To(Equal("m"))
		Expect(got["autoMode"]).To(Equal(map[string]any{"allow": []any{"x"}}))
		Expect(got["hooks"].(map[string]any)["PreToolUse"]).To(HaveLen(1))
	})

	It("backs up the old settings.json once and is a no-op on re-pull", func() {
		local := filepath.Join(home, ".claude", "settings.json")
		Expect(os.MkdirAll(filepath.Dir(local), 0o755)).To(Succeed())
		Expect(os.WriteFile(local, []byte(`{"model":"old"}`), 0o644)).To(Succeed())

		Expect(components.NewClaude(opts).Pull()).To(Succeed())
		Expect(components.NewClaude(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(opts.BackupRoot, "claude", "v1", "settings.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal(`{"model":"old"}`))
		Expect(filepath.Join(opts.BackupRoot, "claude", "v2")).NotTo(BeADirectory())

		var got map[string]any
		b, err = os.ReadFile(local)
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Unmarshal(b, &got)).To(Succeed())
		Expect(got["hooks"].(map[string]any)["PreToolUse"]).To(HaveLen(1))
	})

	It("skips the hook and the settings when disabled in config", func() {
		off := false
		opts.Claude = config.ClaudeConfig{Hook: &off, Settings: &off}
		Expect(components.NewClaude(opts).Pull()).To(Succeed())

		Expect(filepath.Join(home, ".claude", "hooks", "block-unreviewable-edits.sh")).NotTo(BeAnExistingFile())
		Expect(filepath.Join(home, ".claude", "settings.json")).NotTo(BeAnExistingFile())
		Expect(filepath.Join(home, ".claude", "CLAUDE.md")).To(BeARegularFile())
	})

	It("fails on an unparseable local settings.json instead of overwriting it", func() {
		local := filepath.Join(home, ".claude", "settings.json")
		Expect(os.MkdirAll(filepath.Dir(local), 0o755)).To(Succeed())
		Expect(os.WriteFile(local, []byte("{not json"), 0o644)).To(Succeed())

		Expect(components.NewClaude(opts).Pull()).To(MatchError(ContainSubstring("parsing")))

		b, _ := os.ReadFile(local)
		Expect(string(b)).To(Equal("{not json"))
	})

	It("fails when the repo has no rules dir", func() {
		Expect(os.RemoveAll(filepath.Join(repoRoot, "agents", "rules"))).To(Succeed())

		Expect(components.NewClaude(opts).Pull()).NotTo(Succeed())
	})

	Describe("Push", func() {
		It("fails on an unparseable local settings.json", func() {
			hookLocal := filepath.Join(home, ".claude", "hooks", "block-unreviewable-edits.sh")
			Expect(os.MkdirAll(filepath.Dir(hookLocal), 0o755)).To(Succeed())
			Expect(os.WriteFile(hookLocal, []byte("#!/bin/sh\nexit 0\n"), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte("{not json"), 0o644)).To(Succeed())

			Expect(components.NewClaude(opts).Push()).To(MatchError(ContainSubstring("parsing")))
		})

		It("copies the local hook back to the repo, archiving the repo copy under claude-repo", func() {
			hookLocal := filepath.Join(home, ".claude", "hooks", "block-unreviewable-edits.sh")
			Expect(os.MkdirAll(filepath.Dir(hookLocal), 0o755)).To(Succeed())
			Expect(os.WriteFile(hookLocal, []byte("#!/bin/sh\necho new\n"), 0o755)).To(Succeed())

			Expect(components.NewClaude(opts).Push()).To(Succeed())

			b, err := os.ReadFile(filepath.Join(repoRoot, "claude", "hooks", "block-unreviewable-edits.sh"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal("#!/bin/sh\necho new\n"))
			b, err = os.ReadFile(filepath.Join(opts.BackupRoot, "claude-repo", "v1", "block-unreviewable-edits.sh"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal("#!/bin/sh\nexit 0\n"))
		})

		It("writes only the shareable settings keys back to the repo fragment", func() {
			hookLocal := filepath.Join(home, ".claude", "hooks", "block-unreviewable-edits.sh")
			Expect(os.MkdirAll(filepath.Dir(hookLocal), 0o755)).To(Succeed())
			Expect(os.WriteFile(hookLocal, []byte("#!/bin/sh\nexit 0\n"), 0o755)).To(Succeed())
			local := filepath.Join(home, ".claude", "settings.json")
			Expect(os.WriteFile(local, []byte(`{"model":"new","theme":"light","enabledPlugins":{"p":true},"autoMode":{"allow":["x"]},"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"$HOME/.claude/hooks/block-unreviewable-edits.sh"}]},{"matcher":"Edit","hooks":[{"type":"command","command":"/machine/only.sh"}]}]}}`), 0o644)).To(Succeed())

			Expect(components.NewClaude(opts).Push()).To(Succeed())

			var got map[string]any
			b, err := os.ReadFile(filepath.Join(repoRoot, "claude", "settings.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(json.Unmarshal(b, &got)).To(Succeed())
			Expect(got).To(HaveKeyWithValue("model", "new"))
			Expect(got).To(HaveKeyWithValue("theme", "light"))
			Expect(got).To(HaveKey("enabledPlugins"))
			Expect(got).NotTo(HaveKey("autoMode"))
			Expect(got["hooks"].(map[string]any)["PreToolUse"]).To(HaveLen(1))
			Expect(filepath.Join(opts.BackupRoot, "claude-repo", "v1", "settings.json")).To(BeARegularFile())
		})
	})
})
