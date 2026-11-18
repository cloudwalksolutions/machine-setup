package dotfiles_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func preToolUse(hookName, tool string, input map[string]string) (decision, reason string) {
	if _, err := exec.LookPath("jq"); err != nil {
		Skip("jq not installed")
	}
	payload, err := json.Marshal(map[string]any{"tool_name": tool, "tool_input": input})
	Expect(err).NotTo(HaveOccurred())
	hook := exec.Command("bash", filepath.Join(repoRoot(), "claude", "hooks", hookName))
	hook.Stdin = bytes.NewReader(payload)
	out, err := hook.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(out))
	if len(bytes.TrimSpace(out)) == 0 {
		return "allow", ""
	}
	var result struct {
		HookSpecificOutput struct {
			PermissionDecision       string `json:"permissionDecision"`
			PermissionDecisionReason string `json:"permissionDecisionReason"`
		} `json:"hookSpecificOutput"`
	}
	Expect(json.Unmarshal(out, &result)).To(Succeed(), string(out))
	return result.HookSpecificOutput.PermissionDecision, result.HookSpecificOutput.PermissionDecisionReason
}

var _ = Describe("shipped secret-read guard", func() {
	DescribeTable("denies any tool that names a local secret file",
		func(tool string, input map[string]string, path string) {
			decision, reason := preToolUse("block-secret-reads.sh", tool, input)
			Expect(decision).To(Equal("deny"))
			Expect(reason).To(ContainSubstring(path))
		},
		Entry("Read of an .env file", "Read", map[string]string{"file_path": "src/api/.env"}, "src/api/.env"),
		Entry("grep over an .env file", "Bash", map[string]string{"command": "grep FOO src/api/.env"}, "src/api/.env"),
		Entry("cat of a file under .envs", "Bash", map[string]string{"command": "cat .envs/nonprod"}, ".envs/nonprod"),
		Entry("Read of a file under secrets", "Read", map[string]string{"file_path": "src/api/secrets/token.txt"}, "src/api/secrets/token.txt"),
		Entry("Read of a service-account key", "Read", map[string]string{"file_path": "deploy/gcs-creds.json"}, "deploy/gcs-creds.json"),
		Entry("a script that opens an .env variant", "Bash", map[string]string{"command": `python3 -c "open('.env.local')"`}, ".env.local"),
	)

	DescribeTable("lets everything else through",
		func(tool string, input map[string]string) {
			decision, reason := preToolUse("block-secret-reads.sh", tool, input)
			Expect(decision).To(Equal("allow"), reason)
		},
		Entry("Read of .env.example", "Read", map[string]string{"file_path": "src/api/.env.example"}),
		Entry("grep for the code that reads config", "Bash", map[string]string{"command": "grep -rn os.Getenv src"}),
		Entry("a make target that writes env files without printing them", "Bash", map[string]string{"command": "make env-pull"}),
	)
})
