package dotfiles_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("shipped ask-before-tickets hook", func() {
	DescribeTable("asks the user before a ticket is filed",
		func(tool string, input map[string]string) {
			decision, _ := preToolUse("ask-before-tickets.sh", tool, input)
			Expect(decision).To(Equal("ask"))
		},
		Entry("a Jira work item from the twg CLI", "Bash", map[string]string{"command": "twg jira workitem create --summary x"}),
		Entry("a Jira work item from acli", "Bash", map[string]string{"command": "acli jira workitem create --summary x"}),
		Entry("a GitHub issue, after another command", "Bash", map[string]string{"command": "cd ~/src/app && gh issue create --title x"}),
		Entry("a tracker MCP tool that creates an issue", "mcp__atlassian__createJiraIssue", map[string]string{"summary": "x"}),
	)

	DescribeTable("stays silent for everything else",
		func(tool string, input map[string]string) {
			decision, _ := preToolUse("ask-before-tickets.sh", tool, input)
			Expect(decision).To(Equal("allow"))
		},
		Entry("reading a Jira work item", "Bash", map[string]string{"command": "twg jira workitem get BOAT-1"}),
		Entry("viewing a GitHub issue", "Bash", map[string]string{"command": "gh issue view 1"}),
		Entry("a tracker MCP tool that only reads", "mcp__atlassian__getJiraIssue", map[string]string{"issueIdOrKey": "BOAT-1"}),
	)
})
