# pi with tars

`tars pi init` turns a machine with [pi](https://pi.dev) installed into one running the
`tars` baseline agent with the basics, and configures model providers from a form.

## What init does

1. Asks which packages to install (all pre-checked): `pi-subagents` (agent files + the
   `subagent` tool), `pi-mcp-adapter` (MCP servers), `pi-web-access` (`web_search`,
   `fetch_content`), `bigpowers` (skills pack), `@gotgenes/pi-permission-system`,
   `@plannotator/pi-extension` (plan mode with browser-side plan approval, `/plannotator-review`
   code review and markdown annotation UIs), `@juicesharp/rpiv-todo` (a `todo` tool and a
   live task overlay that survives `/reload` and compaction; `/todos` lists them).
2. Asks whether to remove `gentle-pi` if it is installed. It injects its own persona and
   rewrites `~/.pi/agent` at startup, which conflicts with a curated agent; its files are
   backed up under `~/.local/state/tars/backups/pi/` before removal.
3. Asks which model providers to configure, pre-filled from your existing `models.json`
   and a running `ollama`:
   - **ollama**: `http://localhost:11434/v1`, pick the models to expose.
   - **llama.cpp**: server URL; `pi-llama-cpp` registers the provider from the running
     `llama-server`, pick a model with `/models` inside pi.
   - **OpenAI-compatible endpoint**: name, base URL, model ids, and the env var that
     holds the key. A literal key already in `models.json` is moved into `~/.zshrc_secret`
     and replaced by `$THAT_VAR`.
4. Saves the choices under `pi:` in `~/.config/tars/config.yaml`, pulls the files, installs
   the missing packages, and prints what to do next (usually `exec zsh`).

`TARS_NO_FORM=1 tars pi init` takes every default: all packages, remove gentle-pi, keep the
detected providers, no default model pinned.

## What lands in `~/.pi/agent`

| Path | Source |
|---|---|
| `agents/tars.md` | `pi/agents/tars.md` — tools allowlist, `thinking: high`, TDD/PR contract |
| `prompts/tdd.md`, `plan.md`, `pr.md` | `pi/prompts/` — `/tdd`, `/plan`, `/pr` |
| `extensions/block-unreviewable-edits.ts` | `pi/extensions/` — denies `sed -i`, interpreter writes, redirects onto source files |
| `extensions/pi-permission-system/config.json` | `pi/permissions.json` — deny `rm -rf /`, force-push, writes to `~/.ssh`, `*.env` |
| `AGENTS.md` | rendered from `claude/rules/*.md`, same selection as `tars claude init` |
| `settings.json` | merge: `packages` unioned (fragment first), `defaultThinkingLevel`, your default provider/model |
| `models.json` | rendered from the saved providers; local-only providers are kept |

`tars pull` keeps all of it in sync afterwards; `tars push` carries edits to the agent,
prompts, extension and permissions back into the repo. Providers and `models.json` never
move in either direction.

## Adding MCP servers

`pi-mcp-adapter` reads `~/.pi/agent/mcp.json` (global) and a project's `.mcp.json`
(Claude-compatible). Example:

```json
{ "mcpServers": { "github": { "url": "https://api.githubcopilot.com/mcp",
  "headers": { "Authorization": "Bearer ${GITHUB_TOKEN}" } } } }
```
