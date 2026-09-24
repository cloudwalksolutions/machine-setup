import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

// Deny bash commands that edit files in ways a human cannot read as a diff.
// Same five rules as claude/hooks/block-unreviewable-edits.sh; keep them in sync.

const USE_TOOLS =
  "Use the edit tool (exact string replacement) or the write tool (full file) so the change is reviewable as a diff.";

const START = String.raw`(^|[;&|(\x60]|\s)`;
const SOURCE_EXT =
  "(go|ts|tsx|js|jsx|mjs|cjs|py|rb|java|kt|swift|rs|c|h|cc|cpp|hpp|cs|php|sql|json|ya?ml|toml|md|mdx|feature|gql|graphql|proto|scss|sass|css|html|sh|bash|zsh)";

const sedInPlace = new RegExp(START + String.raw`(g?sed)\s+(-[a-zA-Z]*i|--in-place)`);
const perlRubyInPlace = new RegExp(START + String.raw`(perl|ruby)\s+-[pnlaw]*i`);
const awkInPlace = new RegExp(START + String.raw`(g?awk)\s.*-i\s*inplace`);
const interpreter = new RegExp(START + String.raw`(python3?|node|ruby|perl|deno|bun)(\s|$)`);
const writeApi =
  /(write_text|write_bytes|writeFileSync|appendFileSync|writeFile|fs\.write|File\.write|IO\.write|\.writelines\(|open\([^)]*,\s*['"][wa])/;
const redirectToSource = new RegExp(
  String.raw`(>>?\s*|tee\s+(-a\s+)?)["']?[^\s"'|;&]*\.` + SOURCE_EXT + String.raw`(\s|"|'|$|;|&|\|)`,
);
const scratchPath = /(\/tmp\/|\/scratchpad\/|\/dev\/null)/;

export function denialReason(command: string): string | undefined {
  if (sedInPlace.test(command)) return `Blocked: sed in-place edit. ${USE_TOOLS}`;
  if (perlRubyInPlace.test(command)) return `Blocked: perl/ruby in-place edit. ${USE_TOOLS}`;
  if (awkInPlace.test(command)) return `Blocked: awk in-place edit. ${USE_TOOLS}`;
  if (interpreter.test(command) && writeApi.test(command)) {
    return `Blocked: script that writes files (python/node/ruby heredoc or -e). ${USE_TOOLS}`;
  }
  if (redirectToSource.test(command) && !scratchPath.test(command)) {
    return `Blocked: shell redirect over a source file. ${USE_TOOLS}`;
  }
  return undefined;
}

export default function (pi: ExtensionAPI) {
  pi.on("tool_call", async (event) => {
    if (event.toolName !== "bash") return;
    const command = String((event.input as { command?: string }).command ?? "");
    const reason = denialReason(command);
    if (reason) return { block: true, reason };
  });
}
