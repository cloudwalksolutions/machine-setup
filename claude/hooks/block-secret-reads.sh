#!/usr/bin/env bash
set -uo pipefail

named_paths() {
  jq -r '.tool_input | .file_path, .path, .glob, .command | strings' 2>/dev/null \
    | tr ' \t;|&()<>=,`"'"'" '\n'
}

is_secret() {
  case "/$1" in
    */.envs/* | */secrets/*) return 0 ;;
  esac
  case "${1##*/}" in
    .env.example) return 1 ;;
    .env | .env.* | *-creds.json | *credentials*.json) return 0 ;;
  esac
  return 1
}

while IFS= read -r path; do
  is_secret "$path" || continue
  jq -n --arg reason "Blocked: $path is a local secret file. Learn the config from the code that reads it; if the file needs a change, give the user the exact lines to add." '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: $reason
    }
  }'
  exit 0
done < <(named_paths)
exit 0
