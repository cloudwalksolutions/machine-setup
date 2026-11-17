#!/usr/bin/env bash
set -uo pipefail

input="$(cat)"
[ "$(printf '%s' "$input" | jq -r '.stop_hook_active // false')" = "true" ] && exit 0
url="$(gh pr view --json state,url 2>/dev/null | jq -r 'select(.state == "OPEN") | .url')"
[ -z "$url" ] && exit 0

last_message="$(printf '%s' "$input" | jq -r '.last_assistant_message // empty')"
[[ "$last_message" == *"$url"* ]] && exit 0

echo "This branch has an open PR; end the turn with its link on its own line: $url" >&2
exit 2
