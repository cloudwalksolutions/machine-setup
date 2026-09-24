# pi with tars

`tars pi init` turns a machine with [pi](https://pi.dev) installed into one running the
`tars` baseline agent with the basics, and configures model providers from a form.

## What init does

1. Asks which packages to install (all pre-checked): `pi-subagents` (agent files + the
   `subagent` tool), `pi-mcp-adapter` (MCP servers), `pi-web-access` (`web_search`,
   `fetch_content`), `bigpowers` (skills pack), `@gotgenes/pi-permission-system`,
   `@plannotator/pi-extension` (plan mode with browser-side plan approval, `/plannotator-review`
   code review and markdown annotation UIs), `@juicesharp/rpiv-todo` (a `todo` tool and a
   live task overlay that survives `/reload` and compaction; `/todos` lists them), `pi-lens`
   (LSP diagnostics, linters, formatters and type-checking as the agent edits, `symbol_search`,
   `/lens-map` dependency map; needs Node ≥ 22.19), `@dietrichgebert/ponytail` (YAGNI decision
   ladder: reuse, stdlib and native features before new code; `/ponytail-review` flags
   over-engineering in diffs).
2. If a local `ollama` is installed, asks which of its models (`ollama list`) to expose to
   pi and which one is the default. That is the only provider tars manages: llama.cpp,
   remote endpoints and API keys are configured in pi itself (`/login`, `/models`,
   `~/.pi/agent/models.json`) and are left untouched.
3. Saves the choices under `pi:` in `~/.config/tars/config.yaml`, pulls the files, and
   installs the missing packages.

`TARS_NO_FORM=1 tars pi init` takes every default: all packages, every ollama model, no
default pinned.

Packages you no longer want are yours to remove (`pi remove npm:<name>`); tars never
uninstalls anything.

## What lands in `~/.pi/agent`

| Path | Source |
|---|---|
| `agents/tars.md` | `pi/agents/tars.md` — tools allowlist, thinking off, TDD/PR contract |
| `prompts/tdd.md`, `plan.md`, `pr.md` | `pi/prompts/` — `/tdd`, `/plan`, `/pr` |
| `extensions/block-unreviewable-edits.ts` | `pi/extensions/` — denies `sed -i`, interpreter writes, redirects onto source files |
| `extensions/pi-permission-system/config.json` | `pi/permissions.json` — deny `rm -rf /`, force-push, writes to `~/.ssh`, `*.env` |
| `AGENTS.md` | rendered from `claude/rules/*.md`, same selection as `tars claude init` |
| `settings.json` | merge: `packages` unioned (fragment first), `defaultThinkingLevel: off`, your default ollama model |
| `models.json` | the `ollama` provider with the chosen models; every other provider is kept as is |

`tars pull` keeps all of it in sync afterwards; `tars push` carries edits to the agent,
prompts, extension and permissions back into the repo. `models.json` never moves in
either direction.

## Adding MCP servers

`pi-mcp-adapter` reads `~/.pi/agent/mcp.json` (global) and a project's `.mcp.json`
(Claude-compatible). Example:

```json
{ "mcpServers": { "github": { "url": "https://api.githubcopilot.com/mcp",
  "headers": { "Authorization": "Bearer ${GITHUB_TOKEN}" } } } }
```
