package dotfiles_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type bashCall struct{ command, output string }

type stopRun struct {
	calls          []bashCall
	states         map[string]string
	lastMessage    string
	stopHookActive bool
}

func transcript(calls []bashCall) string {
	var lines []string
	for i, call := range calls {
		id := fmt.Sprintf("toolu_%d", i)
		use, err := json.Marshal(map[string]any{"type": "assistant", "message": map[string]any{"content": []any{
			map[string]any{"type": "tool_use", "id": id, "name": "Bash", "input": map[string]string{"command": call.command}},
		}}})
		Expect(err).NotTo(HaveOccurred())
		result, err := json.Marshal(map[string]any{"type": "user", "message": map[string]any{"content": []any{
			map[string]any{"type": "tool_result", "tool_use_id": id, "content": call.output},
		}}})
		Expect(err).NotTo(HaveOccurred())
		lines = append(lines, string(use), string(result))
	}
	return strings.Join(lines, "\n") + "\n"
}

var _ = Describe("shipped PR-link Stop hook", func() {
	stop := func(run stopRun) (int, string) {
		if _, err := exec.LookPath("jq"); err != nil {
			Skip("jq not installed")
		}
		dir := GinkgoT().TempDir()
		fakeGh := "#!/bin/sh\nprintf '{\"state\":\"%s\",\"url\":\"%s\"}' \"$(cat \"$(dirname \"$0\")/state-${3##*/}\")\" \"$3\"\n"
		Expect(os.WriteFile(filepath.Join(dir, "gh"), []byte(fakeGh), 0o755)).To(Succeed())
		for number, state := range run.states {
			Expect(os.WriteFile(filepath.Join(dir, "state-"+number), []byte(state), 0o644)).To(Succeed())
		}
		transcriptPath := filepath.Join(dir, "transcript.jsonl")
		Expect(os.WriteFile(transcriptPath, []byte(transcript(run.calls)), 0o644)).To(Succeed())
		body, err := json.Marshal(map[string]any{
			"transcript_path":        transcriptPath,
			"last_assistant_message": run.lastMessage,
			"stop_hook_active":       run.stopHookActive,
		})
		Expect(err).NotTo(HaveOccurred())

		hook := exec.Command("bash", filepath.Join(repoRoot(), "claude", "hooks", "require-pr-link.sh"))
		hook.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		hook.Stdin = bytes.NewReader(body)
		var stderr bytes.Buffer
		hook.Stderr = &stderr
		err = hook.Run()
		if exit, ok := errors.AsType[*exec.ExitError](err); ok {
			return exit.ExitCode(), stderr.String()
		}
		Expect(err).NotTo(HaveOccurred())
		return 0, stderr.String()
	}

	created := func(url string) bashCall {
		return bashCall{command: "cd ~/src/other-repo && gh pr create --fill", output: url}
	}

	It("blocks the stop when a PR the session created, in any repo, is open and the reply omits its link", func() {
		code, reason := stop(stopRun{
			calls:       []bashCall{created("https://github.com/o/other/pull/7")},
			states:      map[string]string{"7": "OPEN"},
			lastMessage: "All done.",
		})

		Expect(code).To(Equal(2))
		Expect(reason).To(ContainSubstring("https://github.com/o/other/pull/7"))
	})

	DescribeTable("lets the stop through",
		func(run stopRun) {
			code, reason := stop(run)

			Expect(code).To(BeZero(), reason)
		},
		Entry("when the reply carries the link", stopRun{
			calls:       []bashCall{created("https://github.com/o/other/pull/7")},
			states:      map[string]string{"7": "OPEN"},
			lastMessage: "https://github.com/o/other/pull/7 (open)",
		}),
		Entry("when the PR it created has merged", stopRun{
			calls:       []bashCall{created("https://github.com/o/other/pull/7")},
			states:      map[string]string{"7": "MERGED"},
			lastMessage: "All done.",
		}),
		Entry("when an open PR only came up in a lookup, not from gh pr create", stopRun{
			calls:       []bashCall{{command: "gh pr view 39 --json url -q .url", output: "https://github.com/o/r/pull/39"}},
			states:      map[string]string{"39": "OPEN"},
			lastMessage: "All done.",
		}),
		Entry("when it already blocked this stop once, so the turn cannot loop", stopRun{
			calls:          []bashCall{created("https://github.com/o/other/pull/7")},
			states:         map[string]string{"7": "OPEN"},
			lastMessage:    "All done.",
			stopHookActive: true,
		}),
	)
})
