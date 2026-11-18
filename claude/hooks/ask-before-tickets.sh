#!/usr/bin/env bash
set -uo pipefail

input="$(cat)"
tool="$(printf '%s' "$input" | jq -r '.tool_name // empty' 2>/dev/null)"
command_line="$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)"

files_ticket() {
  printf '%s' "$tool" | grep -Eiq '^mcp__.*create.*(issue|ticket|workitem|task)' \
    || printf '%s' "$command_line" | grep -Eq '(^|[;&|(][ 	]*)((twg|acli)[ 	]+jira[ 	]+workitem|gh[ 	]+issue)[ 	]+create'
}

if files_ticket; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "ask",
      permissionDecisionReason: "Filing a ticket: the user decides whether it is filed."
    }
  }'
fi
exit 0
