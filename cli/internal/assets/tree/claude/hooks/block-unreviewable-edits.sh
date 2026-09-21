#!/usr/bin/env bash
# Deny Bash commands that edit files in ways a human cannot read as a diff:
# in-place stream editors, interpreter one-liners/heredocs that write files, and
# redirects clobbering source files. Claude must use the Edit or Write tool instead.
#
# Reads the PreToolUse payload on stdin, prints a deny decision, always exits 0.
set -uo pipefail

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

matches() { printf '%s' "$command_line" | grep -Eq "$1"; }

USE_TOOLS="Use the Edit tool (exact string replacement) or the Write tool (full file) so the change is reviewable as a diff."

# sed -i / -i.bak / --in-place. Reading with sed (no -i) stays allowed.
if matches '(^|[;&|(`]|[[:space:]])(g?sed)[[:space:]]+(-[a-zA-Z]*i|--in-place)'; then
  deny "Blocked: sed in-place edit. $USE_TOOLS"
fi

# perl/ruby -i, -pi, -ni, -pi.bak (i only after p/n/l/a/w so -Ilib is not caught).
if matches '(^|[;&|(`]|[[:space:]])(perl|ruby)[[:space:]]+-[pnlaw]*i'; then
  deny "Blocked: in-place edit via $(printf '%s' "$command_line" | grep -Eo '(perl|ruby)' | head -1). $USE_TOOLS"
fi

# gawk -i inplace
if matches '(^|[;&|(`]|[[:space:]])(g?awk)[[:space:]].*-i[[:space:]]*inplace'; then
  deny "Blocked: awk in-place edit. $USE_TOOLS"
fi

# Interpreter that writes files: python/node/ruby/perl/deno/bun + a write call.
# Reading (json.load(open(p))) is untouched because no write mode/API appears.
if matches '(^|[;&|(`]|[[:space:]])(python3?|node|ruby|perl|deno|bun)([[:space:]]|$)' \
  && matches '(write_text|write_bytes|writeFileSync|appendFileSync|writeFile|fs\.write|File\.write|IO\.write|\.writelines\(|open\([^)]*,[[:space:]]*['"'"'"][wa])'; then
  deny "Blocked: script that writes files (python/node/ruby heredoc or -e). $USE_TOOLS"
fi

# Redirect or tee onto a source file. /tmp and scratchpad paths stay allowed.
SOURCE_EXT='(go|ts|tsx|js|jsx|mjs|cjs|py|rb|java|kt|swift|rs|c|h|cc|cpp|hpp|cs|php|sql|json|ya?ml|toml|md|mdx|feature|gql|graphql|proto|scss|sass|css|html|sh|bash|zsh)'
if matches "(>>?[[:space:]]*|tee[[:space:]]+(-a[[:space:]]+)?)[\"']?[^[:space:]\"'|;&]*\.${SOURCE_EXT}([[:space:]]|\"|'|\$|;|&|\|)" \
  && ! matches '(/tmp/|/scratchpad/|/dev/null)'; then
  deny "Blocked: shell redirect over a source file. $USE_TOOLS"
fi

exit 0
