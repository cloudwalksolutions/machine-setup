#!/usr/bin/env bash
set -uo pipefail

rules="$HOME/.claude/edit-guard.json"
if [ ! -f "$rules" ]; then
  echo "edit guard rules missing: $rules (run tars pull)" >&2
  exit 2
fi

command_line="$(jq -r '.tool_input.command // empty' 2>/dev/null)"
[ -z "$command_line" ] && exit 0

deny() {
  jq -n --arg reason "$1" '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: $reason
    }
  }'
  exit 0
}

matches() { printf '%s' "$command_line" | grep -Eq -- "$1"; }

USE_TOOLS="Use the Edit tool (exact string replacement) or the Write tool (full file) so the change is reviewable as a diff."

while IFS= read -r -d '' reason && IFS= read -r -d '' match \
  && IFS= read -r -d '' and && IFS= read -r -d '' unless; do
  if matches "$match" \
    && { [ -z "$and" ] || matches "$and"; } \
    && { [ -z "$unless" ] || ! matches "$unless"; }; then
    deny "Blocked: $reason. $USE_TOOLS"
  fi
done < <(jq -j '.[] | .reason, "\u0000", .match, "\u0000", (.and // ""), "\u0000", (.unless // ""), "\u0000"' "$rules")

exit 0
