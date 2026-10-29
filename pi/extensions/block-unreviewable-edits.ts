import { readFileSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

// Deny bash commands that edit files in ways a human cannot read as a diff.
// The rules live in ~/.pi/agent/edit-guard.json (pulled from agents/edit-guard.json,
// shared with Claude Code's hook); this extension only evaluates them.

export type Rule = { reason: string; match: string; and?: string; unless?: string };

const USE_TOOLS =
  "Use the edit tool (exact string replacement) or the write tool (full file) so the change is reviewable as a diff.";

const matches = (pattern: string, command: string) => new RegExp(pattern, "m").test(command);

export function denialReason(command: string, rules: Rule[]): string | undefined {
  for (const rule of rules) {
    if (
      matches(rule.match, command) &&
      (!rule.and || matches(rule.and, command)) &&
      (!rule.unless || !matches(rule.unless, command))
    ) {
      return `Blocked: ${rule.reason}. ${USE_TOOLS}`;
    }
  }
  return undefined;
}

export function loadRules(path = join(homedir(), ".pi", "agent", "edit-guard.json")): Rule[] {
  return JSON.parse(readFileSync(path, "utf8"));
}

export default function (pi: ExtensionAPI) {
  const rules = loadRules();
  pi.on("tool_call", async (event) => {
    if (event.toolName !== "bash") return;
    const command = String((event.input as { command?: string }).command ?? "");
    const reason = denialReason(command, rules);
    if (reason) return { block: true, reason };
  });
}
