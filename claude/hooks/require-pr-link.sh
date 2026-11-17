#!/usr/bin/env bash
set -uo pipefail

input="$(cat)"
[ "$(printf '%s' "$input" | jq -r '.stop_hook_active // false')" = "true" ] && exit 0
transcript="$(printf '%s' "$input" | jq -r '.transcript_path // empty')"
[ -f "$transcript" ] || exit 0
last_message="$(printf '%s' "$input" | jq -r '.last_assistant_message // empty')"

created_prs() {
  jq -rs '
    [ .[] | select(.type == "assistant") | .message.content[]?
      | select(.type == "tool_use" and ((.input.command? // "") | test("(^|[;&|(]\\s*)gh pr create")))
      | .id ] as $ids
    | .[] | select(.type == "user") | .message.content[]?
    | select(.type == "tool_result" and (.tool_use_id as $id | $ids | index($id)))
    | .content | if type == "array" then map(.text? // "") | join("\n") else tostring end
    | scan("https://github\\.com/[^/ \"]+/[^/ \"]+/pull/[0-9]+")
  ' "$transcript" 2>/dev/null | sort -u
}

missing=()
while IFS= read -r url; do
  [ -z "$url" ] && continue
  [[ "$last_message" == *"$url"* ]] && continue
  state="$(gh pr view "$url" --json state,url 2>/dev/null | jq -r '.state')"
  [ "$state" = "OPEN" ] && missing+=("$url")
done < <(created_prs)

[ ${#missing[@]} -eq 0 ] && exit 0

echo "These PRs from this session are still open; end the turn with each link on its own line and its state: ${missing[*]}" >&2
exit 2
