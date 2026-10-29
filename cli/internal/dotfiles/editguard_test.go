package dotfiles_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("shipped edit guards", func() {
	rulesTable := func() []byte {
		return []byte(read(filepath.Join(repoRoot(), "agents", "edit-guard.json")))
	}

	hookDecision := func(command string) string {
		if _, err := exec.LookPath("jq"); err != nil {
			Skip("jq not installed")
		}
		home := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(home, ".claude"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".claude", "edit-guard.json"), rulesTable(), 0o644)).To(Succeed())
		payload, err := json.Marshal(map[string]any{"tool_input": map[string]string{"command": command}})
		Expect(err).NotTo(HaveOccurred())

		hook := exec.Command("bash", filepath.Join(repoRoot(), "claude", "hooks", "block-unreviewable-edits.sh"))
		hook.Env = append(os.Environ(), "HOME="+home)
		hook.Stdin = bytes.NewReader(payload)
		out, err := hook.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
		if len(bytes.TrimSpace(out)) == 0 {
			return "allow"
		}
		var decision struct {
			HookSpecificOutput struct {
				PermissionDecision       string `json:"permissionDecision"`
				PermissionDecisionReason string `json:"permissionDecisionReason"`
			} `json:"hookSpecificOutput"`
		}
		Expect(json.Unmarshal(out, &decision)).To(Succeed(), string(out))
		Expect(decision.HookSpecificOutput.PermissionDecision).To(Equal("deny"))
		return decision.HookSpecificOutput.PermissionDecisionReason
	}

	extensionDecision := func(command string) string {
		node, err := exec.LookPath("node")
		if err != nil {
			Skip("node not installed")
		}
		const script = `const [extension, rules, command] = process.argv.slice(1);
const { denialReason } = await import(extension);
const { readFileSync } = await import("node:fs");
process.stdout.write(denialReason(command, JSON.parse(readFileSync(rules, "utf8"))) ?? "allow");`
		run := exec.Command(node, "--experimental-strip-types", "--no-warnings", "--input-type=module", "-e", script, "--",
			filepath.Join(repoRoot(), "pi", "extensions", "block-unreviewable-edits.ts"),
			filepath.Join(repoRoot(), "agents", "edit-guard.json"),
			command)
		out, err := run.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
		return string(out)
	}

	decision := func(want string) types.GomegaMatcher {
		if want == "allow" {
			return Equal("allow")
		}
		return HavePrefix("Blocked: " + want + ".")
	}

	DescribeTable("deny the same commands for the same reason and allow the rest, from the one shared table",
		func(command, want string) {
			Expect(hookDecision(command)).To(decision(want))
			Expect(extensionDecision(command)).To(decision(want))
		},
		Entry("sed in place", "sed -i 's/a/b/' main.go", "sed in-place edit"),
		Entry("sed reading", "sed -n '1,5p' main.go", "allow"),
		Entry("perl in place", "perl -pi -e 's/a/b/' x.rb", "perl/ruby in-place edit"),
		Entry("perl with an include path", "perl -Ilib script.pl", "allow"),
		Entry("gawk in place", "gawk -i inplace '{print}' f.txt", "awk in-place edit"),
		Entry("python heredoc that writes", "python3 - <<'EOF'\nopen('x.py', 'w').write('')\nEOF", "script that writes files"),
		Entry("python that only reads", `python3 -c "import json; json.load(open('x.json'))"`, "allow"),
		Entry("redirect onto a source file", "echo hi > main.go", "shell redirect over a source file"),
		Entry("redirect into /tmp", "echo hi > /tmp/out.go", "allow"),
		Entry("tee onto a source file", "tee -a notes.md", "shell redirect over a source file"),
		Entry("redirect to /dev/null", "go test ./... > /dev/null", "allow"),
	)
})
