package dotfiles_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const openPR = `{"state":"OPEN","url":"https://github.com/o/r/pull/7"}`

var _ = Describe("shipped PR-link Stop hook", func() {
	stop := func(payload map[string]any, ghOutput string) (int, string) {
		if _, err := exec.LookPath("jq"); err != nil {
			Skip("jq not installed")
		}
		bin := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nprintf '%s' '"+ghOutput+"'\n"), 0o755)).To(Succeed())
		body, err := json.Marshal(payload)
		Expect(err).NotTo(HaveOccurred())

		hook := exec.Command("bash", filepath.Join(repoRoot(), "claude", "hooks", "require-pr-link.sh"))
		hook.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))
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

	It("blocks the stop when the branch has an open PR and the last message omits its link", func() {
		code, reason := stop(map[string]any{"last_assistant_message": "All done."}, openPR)

		Expect(code).To(Equal(2))
		Expect(reason).To(ContainSubstring("https://github.com/o/r/pull/7"))
	})

	DescribeTable("lets the stop through",
		func(payload map[string]any, ghOutput string) {
			code, reason := stop(payload, ghOutput)

			Expect(code).To(BeZero(), reason)
		},
		Entry("when the last message carries the link",
			map[string]any{"last_assistant_message": "https://github.com/o/r/pull/7 (open)"}, openPR),
		Entry("when the branch's PR is merged",
			map[string]any{"last_assistant_message": "All done."}, `{"state":"MERGED","url":"https://github.com/o/r/pull/7"}`),
		Entry("when it already blocked this stop once, so the turn cannot loop",
			map[string]any{"stop_hook_active": true, "last_assistant_message": "All done."}, openPR),
	)
})
