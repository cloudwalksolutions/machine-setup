# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

## Repository Purpose

A dev-machine provisioning tool for macOS (primary) and Linux (incl. shared bastions). It goes
from a fresh machine to a production-ready dev environment: dev tools, dotfiles
(Neovim, Zsh, Byobu, Vim), fonts, and terminal settings, plus a declarative byobu
session manager. The user-facing CLI is **`tars`** (Go, in `cli/`) with seven verbs:
`setup`, `pull`, `push`, `sessions`, `profiles`, `claude`, `pi`.

## Core Philosophy

- **One command to provision**: `tars setup` installs packages (brew on macOS,
  apt/tarball on Linux), oh-my-zsh, Powerlevel10k, then applies all configs.
  Day-to-day: `pull`/`push` sync configs, `sessions` opens byobu workspaces, `profiles`
  switches git/GitHub/SSH identity per project dir.
- **Bidirectional sync**: `tars pull` applies repo configs to the machine (no installs,
  no network — safe to re-run); `tars push` copies local edits back into the repo.
- **Versioned backups**: every overwrite is archived first under
  `~/.local/state/tars/backups/<component>/vN/` (`<component>-repo/vN/` for push;
  override with `TARS_BACKUP_ROOT`). Per-user, never in the repo clone, never
  auto-deleted. Identical files are skipped (no backup, no copy) so re-runs are no-ops.
  `pull`/`push` exit non-zero when any component fails; `setup` tolerates pull failures.
- **Self-contained Go**: the CLI reimplements all logic natively. It NEVER shells out to
  repo scripts. (The old `scripts/` bash tooling has been removed; `cli/` is the source
  of truth.)
- **Personalization**: Zsh sources optional `~/.zshrc_secret`, the active profile's
  `~/.config/tars/profiles/<alias>.env`, and `~/.zshrc_funcs`; the login shell sources
  `~/.zprofile_local`. All are git-ignored, seeded once, never overwritten. Keep shared
  configs machine-agnostic — no hardcoded personal paths, no niche tools.

## Architecture

```
├── cli/                         # the `tars` Go CLI (module lives here)
│   ├── main.go                  # entrypoint; version/commit/date injected via ldflags
│   ├── cmd/                     # cobra commands: setup, pull, push, sessions, profiles (+ root)
│   │   ├── setup.go             # Setup orchestrator + SequentialPuller (DI, testable)
│   │   ├── pull.go / push.go    # apply / capture configs; SequentialPusher
│   │   ├── sessions.go          # byobu sessions command group (aliases s/by)
│   │   ├── profiles.go          # account identities command group (alias p)
│   │   ├── backup.go            # BackupRoot: ~/.local/state/tars/backups (+ env override)
│   │   ├── resolve.go           # ResolveRepo: clone discovery → embedded-assets fallback
│   └── internal/
│       ├── components/          # per-tool Pull/Push: vim, zsh, byobu, nvim, fonts, terminal, profiles, claude, pi
│       ├── sessions/            # declarative byobu sessions: yaml config + idempotent launcher
│       ├── profiles/            # account identities: yaml config, gitconfig/env rendering, active marker
│       ├── dotfiles/            # specs only: lint the shipped zsh files + terminal font string
│       ├── assets/              # dotfiles embedded in the binary (tree/ mirror; `make sync-assets`)
│       ├── fsutil/              # Backup + SafeCopy (versioned, idempotent)
│       ├── paths/               # repo→local file mappings (ForOS: OS-aware)
│       ├── repo/                # repo-root discovery (markers: cli/go.mod + nvim/)
│       ├── pkg/                 # installable dev tools (brew/apt/rvm) + registry
│       ├── forms/               # huh TUI (honors TARS_NO_FORM=1)
│       ├── shell/               # oh-my-zsh / powerlevel10k installers
│       └── config/              # persisted YAML config
├── nvim/  zsh/  byobu/  vim/     # the dotfiles tars manages
├── monokai.lua                  # nvim colorscheme shipped at root (also embedded)
├── claude/                      # Claude Code: settings fragment, hooks/, rules/ (→ ~/.claude/CLAUDE.md), templates/
├── pi/                          # pi coding agent: settings fragment, agents/, prompts/, extensions/, permissions.json
├── fonts/                       # Hack Nerd Font files
├── terminal/                    # font string + Terminal.app profile (iTerm2/Terminal.app)
├── vhs/                         # README demo tapes + gh stand-in (`make demos`)
├── test/e2e/Dockerfile          # Docker E2E (see Testing Strategy)
├── docs/releasing.md            # maintainer release notes
├── install.sh                   # curl|sh installer (release binary → ~/.local/bin)
├── Makefile                     # dev targets: build/lint/test layers, sync-assets, e2e, demos
└── .goreleaser.yaml, .github/   # release (tag v*) + CI + vhs re-recording
```

The CLI locates the repo root by walking up for `cli/go.mod` + `nvim/`, or via the
`TARS_REPO` env var (`cli/internal/repo`). With no clone at all, `cmd.ResolveRepo`
falls back to the dotfiles embedded in the binary (`cli/internal/assets`), materialized
under `~/.local/share/tars/repo` — so a brew-installed `tars` needs no clone. The
embedded mirror is refreshed with `make sync-assets`; a drift-guard spec fails when a
dotfile changes without re-syncing. `push` still requires a real clone.

## Testing Strategy

**Framework: Ginkgo v2 + Gomega.** Each package has a `*_suite_test.go` runner and
`*_test.go` specs. The Makefile maps the layers to targets: `make unit` (fast,
airgapped), `make integration` (external deps), `make e2e` (Docker), `make test`
(unit + integration), `make check` (lint + build + test). CI runs `go test -race`.

Tests are layered:

1. **Unit** (`make unit` = `go test ./...`) — the fast, airgapped bulk, no external
   deps or network. Two kinds:
   - Logic specs (`internal/...`) with `GinkgoT().TempDir()` fake repos/HOMEs — component
     Pull/Push, `fsutil` backup/copy, `paths` OS-awareness, the `pkg` registry.
   - Command specs (`cmd/`) that construct `Setup`/`SequentialPuller`/`SequentialPusher`/
     `Sessions`/`Profiles` with **spy collaborators** and assert orchestration (order, failure-tolerance),
     never touching the real machine.
2. **Integration** (`make integration`) — real external deps: brew installers gated by
   `INTEGRATION=1` (installs/removes `hello`) plus the Neovim config tests (real `nvim`).
   Off by default in `go test`.
3. **End-to-end** (`make e2e`) — `test/e2e/Dockerfile`: builds `tars` against a
   **root-owned, read-only** repo clone shared by two non-root users, plus a third user
   with no clone (embedded-assets path) — **no apt installs, no sudo, no network at
   runtime**. Asserts the file-producing components land (vim, zsh, byobu, nvim, fonts,
   claude; terminal and profiles are no-ops there and pull exiting 0 covers them), incl.
   the exec bit on `byobu/bin` and `~/.claude/hooks` and the settings.json merge keeping
   local keys, the
   pulled shell configs parse (`zsh -ic` / `bash -lc`), backups version under each
   user's `$HOME`, re-pulls are idempotent (incl. nvim), and a failing pull exits
   non-zero. Assertions are `RUN` lines, so a failure fails `docker build`. Runs in CI
   on every PR on amd64 and arm64.

**The seam pattern (critical).** Unit tests must never touch the real system (plists,
`sudo cp`, network, `/Library/Fonts`). Anything that does is injected behind a seam so a
spec can drive a fake and still fail red-first:

- **Function-typed fields** for side-effecting ops, e.g. `Fonts.CopyFn(src,dst)` (real:
  `sudo cp`), and `Terminal.CurrentFontFn / SetFontFn / IsRunningFn / DefaultProfileFn /
  ApplyFn / ExportFn` (real: PlistBuddy / `pgrep` / `open` / `defaults`). Tests assign
  fakes; the real impl is left untested by unit tests and validated via the E2E and
  manual runs.
- **Dry-run seam**: every component writes through `Options.copier()` → `fsutil.Copier`;
  with `Options.DryRun` it reports `would create / would overwrite / unchanged / would
  remove` instead of touching disk (`tars pull --dry-run`). Specs assert on the report.
- **OS override constructors**: `NewFontsForOS(opts, goos)`, `NewTerminalForOS(opts, goos)`
  so darwin-only paths are exercised on any host. CI also runs a `macos-latest` matrix leg
  so darwin-only code compiles and its unit tests run for real.
- **Path/env overrides**: `Fonts.LocalOverride`, `TARS_REPO`,
  `TARS_NO_FORM=1` (skips the TUIs), `TARS_CONFIG_PATH`,
  `TARS_SESSIONS_PATH`, `TARS_PROFILES_PATH`, `TARS_BACKUP_ROOT`,
  `Profiles.ConfigPath` (component), and
  `UPDATE_GOLDEN=1` (regenerates `components/testdata/pull_manifest.golden`; review the diff).

**Backup-safety invariant (must hold, is tested).** Any destructive write (e.g.
`os.RemoveAll` in `nvim` Pull/Push) MUST call `fsutil.Backup(...)` first AND error-check
it before removing — a backup failure must abort before any data is lost. `SafeCopy`
enforces backup-before-overwrite and short-circuits on byte-identical content.

**Coverage & mutation testing.** CI enforces the total-coverage threshold in
`cli/.testcoverage.yml` (ratchet it up when coverage grows, never down). Test
*quality* is spot-checked with `go-mutesting` (avito-tech): run
`go-mutesting ./internal/<pkg>/` from `cli/` — it mutates files IN PLACE while
running, so never edit or test concurrently; surviving mutants point at missing
assertions. Production seams (`Default*` runners, `forms/`, plist helpers) stay
untested by design — don't chase 100%.

**Neovim config is tested via Lua, not Go**: `make test-nvim` runs headless smoke
tests (`nvim/tests/smoke_test.lua`); `make health-nvim` runs checkhealth. `make test-nvim`
is also run as part of `make integration`.

**CI** (`.github/workflows/ci.yml`, on PRs) is the documentation of what "green" means;
its steps are explicit commands, not Makefile recipes. Eight checks:
`test (ubuntu-latest)` / `test (macos-latest)` (`go vet`, `go build`, `go test -race`),
`lint` (golangci-lint incl. the gofmt formatter, then `go mod tidy -diff`), `coverage`
(gate from `cli/.testcoverage.yml`), `e2e-linux (ubuntu-latest)` / `e2e-linux (ubuntu-24.04-arm)`
(the Docker build), `nvim` (headless Lua smoke tests against the repo's `nvim/` via
`XDG_CONFIG_HOME`), and `goreleaser` (`goreleaser check`). Reproduce locally from `cli/` with the
same commands before handing work over; `gh pr checks <n> --watch` is the final gate.
`.github/workflows/vhs.yml` re-records the README GIFs via PR when anything under `vhs/`
other than the GIFs changes on `main`.

## Development Patterns (TDD)

**Strict red → green, one spec at a time.** This is non-negotiable for `cli/` code:

1. Write ONE failing `It(...)` for the next small behavior.
2. Run just that package (`go test ./internal/<pkg>/`), and SEE it fail (compile error or
   assertion). Do not skip observing red.
3. Write the MINIMAL code to make it green.
4. Refactor if needed, keeping green. Then move to the next spec.

Do NOT batch-write many specs or implement ahead of a failing test. Match existing
Ginkgo/Gomega style and the seam pattern above.

**Adding a new dotfile component** (Go, TDD each step):

1. Add its paths to `paths.go` (`<Comp>Paths` struct + wire into `ForOS`, OS-aware if
   needed) — spec first in `paths_test.go`.
2. Implement `cli/internal/components/<comp>.go` with `Name()` + `Pull()` (and `Push()`
   if it's pushable). Use `opts.copier().SafeCopy(src, dst, "<comp>", opts.BackupRoot)`
   (or `.SafeWrite` for rendered content) so backup-before-overwrite, idempotent skip,
   and `--dry-run` all come for free; for push, use component name `"<comp>-repo"`.
   Put any system side-effects behind a function-typed seam.
3. Register in `AllPullable` / `AllPushable` (`component.go`) — spec asserts membership;
   then `UPDATE_GOLDEN=1 go test ./internal/components/` and review the golden diff.
4. `gofmt`, then `go test ./... && golangci-lint run ./...` green.

**Adding a global Claude rule**: drop a short `NN-slug.md` (a `##` heading plus a few
lines) into `claude/rules/`, run `make sync-assets`. The `claude` component renders the
selected rules into `~/.claude/CLAUDE.md`; `claude/settings.json` holds only the shareable
keys (model, theme, enabledPlugins, the hook entry) in `json.MarshalIndent` key order so
`push` round-trips byte-for-byte. Never sync `autoMode`, `permissions`, or `~/.claude.json`.

**pi**: `pi/` holds the shareable pieces (agent, prompts, the edit-guard extension that
ports `claude/hooks/block-unreviewable-edits.sh` rule for rule, the permission baseline, a
settings fragment with packages). `~/.pi/agent/AGENTS.md` is rendered from `claude/rules/`
(one rule source for both agents). The only provider tars manages is local ollama: `init`
asks which `ollama list` models to expose and the default, stores that under `pi:` in the tars
config, and renders the `ollama` entry of `models.json`; every other provider, key and
llama.cpp setup is pi's own business and is left untouched. Thinking is off by default. Only
`init` shells out to `pi`/`ollama` (behind `Pi.Run`); `pull` writes files only; nothing is ever
uninstalled. Design for an empty machine: detecting what the machine has to pre-fill the form
is fine, code that only migrates one machine's state is not. Add a prompt by dropping
`pi/prompts/<name>.md` and running `make sync-assets`.

**Adding an installable tool**: edit the curated lists in
`cli/internal/pkg/registry.go` (`darwinFormulas` / `darwinCasks` / `darwinTappedFormulas`
/ `linuxAptPackages`, or a dedicated `apt` installable) — registry spec first. The
registry, not any script, is the source of truth for installed tooling.

**Conventions**: keep comments to a single concise line (state the why, not a paragraph).
Never edit files with `sed`/`awk` stream edits — use proper edits and fix at the source.

## Customization and Overrides (git-ignored)

- `~/.zshrc_secret` — API keys, tokens, per-account aliases (template:
  `zsh/zshrc_secret.template`; seeded by `setup` if missing).
- `~/.zprofile_local` — per-machine PATH/env for the login shell (template:
  `zsh/zprofile_local.template`; seeded once by `pull`, never overwritten).
- `~/.zshrc_funcs` — personal shell functions (never pushed back into the repo).
- `~/.config/tars/profiles/<alias>.env` — per-account tokens and aliases, sourced by
  `~/.zshrc` while that profile is active (`tars profiles use`). Seeded once by the
  `profiles` component, never overwritten; the component itself is Pullable only.
  Profiles derive everything path-like by convention: dir `<projects_dir>/<name>`,
  key `~/.ssh/id_rsa.<alias>`; add no explicit dir/key fields.
- Keep shared configs generic; machine-specific bits belong in `*_local` files sourced by
  the canonical config (e.g. `~/.zprofile` sources `~/.zprofile_local`).

## Releasing

Every merge to `main` deploys: `.github/workflows/release.yml` bumps the patch tag, runs
GoReleaser (darwin/linux × amd64/arm64 archives + checksums, GitHub Release, Homebrew tap),
refreshes the coverage badge, and pushes the tag last so a failed run leaves nothing behind.
Docs-only merges (`**.md`, `docs/`, `vhs/`) skip it. CI (`ci.yml`) runs on PRs only. Minor/major
bumps are a manual tag push; see `docs/releasing.md`. Test locally with
`HOMEBREW_TAP_TOKEN=x goreleaser release --snapshot --clean`.

## Important Notes

- **Never commit secrets.** `.gitignore` excludes `.zshrc_secret`, `backups/`, and build
  artifacts — stay vigilant.
- **Backups are versioned** (v1, v2, v3…), live under `~/.local/state/tars/backups`,
  and are never auto-deleted.
- **macOS-first, Linux-supported.** Homebrew, `/Library/Fonts`, iTerm2/Terminal.app
  plists are macOS; Linux gets the full dotfile pull plus apt/tarball installs
  (Neovim tarball, GitHub/GCloud apt repos) suitable for shared bastions.
- **Neovim is the primary editor**; vim is a minimal fallback. Neovim needs Python3,
  Node.js, and language servers (auto-installed via Mason).
- **ALWAYS run Neovim tests**: after ANY change to the Neovim config, run `make test-nvim`
  automatically (don't ask) to validate.
- The user handles all `git` operations themselves — make changes and stop; don't commit,
  branch, or push.
