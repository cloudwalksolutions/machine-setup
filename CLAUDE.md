# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

## Repository Purpose

`tars` (Go, `cmd/` + `internal/`) provisions macOS (primary) and Linux dev machines, incl.
shared bastions: dev tools, dotfiles (Neovim, Zsh, Byobu, Vim), fonts, terminal settings,
coding-agent config, and live byobu sessions, one per project under `projects_dir`. Verbs:
`init` (subcommands `agents`, `claude`, `pi`, `project`), `pull`, `push`, `sessions`, `profiles`.

## Core Philosophy

- **`pull` is safe to re-run**: it applies repo configs to the machine with no installs and no
  network; `push` copies local edits back into the repo. Both exit non-zero when any
  component fails; `init` tolerates pull failures.
- **Versioned backups**: every overwrite or removal is archived first under
  `~/.local/state/tars/backups/<component>/vN/` (`<component>-repo/vN/` for push; override
  with `TARS_BACKUP_ROOT`), never in the clone, never auto-deleted. Byte-identical files are
  skipped, so re-runs are no-ops.
- **Self-contained Go**: the CLI never shells out to repo scripts.
- **Machine-agnostic configs**: no hardcoded personal paths or niche tools; machine-specific
  bits live in the git-ignored files listed under Customization.

The CLI finds the repo root by walking up for `go.mod` + `nvim/`, or via `TARS_REPO`. With
no clone, `cmd.ResolveRepo` falls back to the dotfiles embedded in the binary (`embed.go`,
served by `internal/assets`, materialized under `~/.local/share/tars/repo`), so a
brew-installed `tars` needs no clone. The embed reads the real directories, so every dotfile
exists once and ships on the next build. `push` needs a real clone.

## Testing Strategy

Ginkgo v2 + Gomega; each package has a `*_suite_test.go` runner. Layers:

1. **Unit** (`make unit` = `go test ./...`): airgapped. Logic specs over `GinkgoT().TempDir()`
   fake repos and HOMEs; command specs (`cmd/`) that build the orchestrators with spy
   collaborators and assert order and failure tolerance. Orchestrators never print: they call
   a `report.Reporter` and ask through narrow prompt interfaces that `tui.Prompts` and
   `forms.Headless` both satisfy. TUI specs drive `Model.Update`/`View` with messages, and
   `Start` under a real program via `teatest/v2`; only `run.go` is untested.
2. **Integration** (`make integration`, `INTEGRATION=1`): brew installers, the session
   launcher against real byobu/tmux on an isolated socket, and the Neovim tests.
3. **End-to-end** (`make e2e`): `test/e2e/Dockerfile` pulls against a root-owned, read-only
   clone shared by two non-root users, plus a user with no clone, with no apt, sudo or network
   at runtime. Assertions are `RUN` lines, so a failure fails `docker build`.

**What specs cover.** Specs prove tars's own mechanics: what gets copied, merged, versioned,
backed up and skipped. They don't pin the contents of shipped agent config (pi agent
frontmatter, `settings.json` hook entries, package lists). Those only restate the file, and
what they do to Claude or pi can't be unit-tested, so such a change is proven by the PR's
manual QA plan. Scripts tars ships, like the hooks, are still code and get specs that run them.

**The seam pattern.** Unit tests never touch the real system (plists, network, package
managers, PATH lookups). Anything that does sits behind a seam a spec can fake:

- **Function-typed fields** for side effects, e.g. `pkg.PathProbe{LookPath, Run}` and
  `Terminal.CurrentFontFn / SetFontFn / IsRunningFn / DefaultProfileFn / ApplyFn / ExportFn`.
  The real implementations are covered by the e2e and manual runs.
- **The copier**: every component writes through `Options.copier()` → `fsutil.Copier`. Each
  write it makes or would make is a `fsutil.Change` handed to `Options.Record` and printed:
  `would create / would overwrite (backup vN) / would remove` under `DryRun`,
  `created / overwrote / removed` for real; `Diff` attaches hunks. `SequentialPuller.Record`
  collects them into `components.Plan`s for the review (`tui.Review`), `--diff` and the recap.
  Never write around the copier, or dry-run leaks.
- **OS override constructors**: `NewFontsForOS`, `NewTerminalForOS` exercise darwin paths on any
  host; CI's `macos-latest` leg runs them for real.
- **Path/env overrides**: `TARS_REPO`, `TARS_NO_FORM=1` (plain-text reporter, default answers),
  `TARS_CONFIG_PATH`, `TARS_PROJECTS_PATH`, `TARS_PROFILES_PATH`, `TARS_BACKUP_ROOT`,
  `Fonts.LocalOverride`, `Profiles.ConfigPath`, and `UPDATE_GOLDEN=1` (regenerates
  `components/testdata/pull_manifest.golden`; review the diff).

**Backup-safety invariant (tested).** A destructive write (e.g. `RemoveAll` in `nvim`) calls
`Backup` first and error-checks it, so a failed backup aborts before data is lost. `SafeCopy`
backs up before overwriting and skips byte-identical content.

**Coverage and mutation testing.** CI enforces `.testcoverage.yml` (ratchet it up, never
down). `go-mutesting ./internal/<pkg>/` spot-checks test quality; it mutates files in place,
so never edit or test while it runs. Production seams (`Default*` runners, `forms/`, plist
helpers) stay untested by design.

**Neovim** is tested in Lua: after any change to `nvim/`, run `make test-nvim` without asking.

**CI** (`ci.yml`, on PRs) spells out what green means, as explicit commands: `test`
(ubuntu + macos: `go vet`, `go build`, `go test -race`), `lint` (golangci-lint, then
`go mod tidy -diff`), `coverage`, `e2e-linux` (amd64 + arm64), `nvim`, `goreleaser check`.
`vhs.yml` records every tape (`demo`, `pull`, `init`, `sessions`, `profiles`) and posts the
GIFs as a sticky PR comment; on `main` it re-records the README GIFs via PR.

**Traps**
- Tapes run `tars` in a real terminal, so a change to what a command prints, asks or which
  flags it takes means updating `vhs/*.tape` in the same PR and recording with `make demos`.
  Don't commit regenerated GIFs (`git checkout vhs/*.gif`). A new prompt hangs a tape unless
  it answers or skips it (`tars pull -y`); VHS strings can't escape quotes; `Wait+Screen` only
  sees the visible screen, so after long output use `Wait`.
- `go test` caches results by Go inputs only: after changing a shell script the dotfiles
  specs run, use `go test -count=1`.
- To repeat a suite use `go run github.com/onsi/ginkgo/v2/ginkgo -repeat=N`; Ginkgo refuses
  `go test -count=N` for N > 1.

## Procedures

**Adding a dotfile component**, spec first at each step:
1. Paths in `paths.go` (`<Comp>Paths`, wired into `ForOS`).
2. `internal/components/<comp>.go` with `Name()`, `Pull()` (and `Push()`), writing through
   `opts.copier().SafeCopy(src, dst, "<comp>", opts.BackupRoot)` or `.SafeWrite`; push uses
   `"<comp>-repo"`. System side effects go behind a function-typed seam.
3. Register in `AllPullable` / `AllPushable`, then `UPDATE_GOLDEN=1 go test ./internal/components/`.

**Adding a global agent rule**: drop a short `NN-slug.md` (a `##` heading plus a few lines)
into `agents/rules/`. Every rule not under `agents.skip_rules` lands on the next `tars pull`.
Claude Code gets one file per rule in `~/.claude/rules/tars/`, a folder tars owns, so skipped
rules can be removed without touching the user's own rules; pi gets them rendered into
`~/.pi/agent/AGENTS.md` by `pullRules` in `components/agents.go`.

**The edit guard**: `agents/edit-guard.json` is the one table of denied-command patterns
(`match`, optional `and` and `unless`, each a regex in the subset POSIX ERE and JavaScript
share, so no `\s` or `[[:space:]]`). `pullEditGuard` copies it next to each agent's guard, and
the bash hook and the pi extension only evaluate it; `internal/dotfiles/editguard_test.go`
drives both through the same table (needs `jq` and `node`).

**Claude settings**: `claude/settings.json` holds only the shareable keys (model, theme,
enabledPlugins, hook entries) in `json.MarshalIndent` key order so `push` round-trips
byte-for-byte. Never sync `autoMode`, `permissions`, or `~/.claude.json`.

**pi**: `pi/` holds the agent, prompts, the edit-guard extension, the permission baseline and a
settings fragment with packages. AGENTS.md is the only statement of the rules; the agent file
and prompts say how to run a task and never restate a rule. The only provider tars manages is
local ollama (`init` asks which models to expose and renders the `ollama` entry of
`models.json`); other providers are pi's business. Only `init` shells out to `pi`/`ollama`
(behind `Pi.Run`); nothing is ever uninstalled. Design for an empty machine: detecting state
to pre-fill a form is fine, code that migrates one machine's state is not. Add a prompt by
dropping `pi/prompts/<name>.md`.

**Adding an installable tool**: edit the curated lists in `internal/pkg/registry.go`
(`darwinFormulas`, `darwinCasks`, `darwinTappedFormulas`, `linuxAptPackages`, or a dedicated
apt installable), registry spec first.

## Customization and Overrides (git-ignored)

- `~/.zshrc_secret`: API keys, tokens, per-account aliases (template
  `zsh/zshrc_secret.template`; seeded by `init` if missing).
- `~/.zprofile_local`: per-machine PATH/env for the login shell (seeded once by `pull`).
- `~/.zshrc_funcs`: personal shell functions, never pushed.
- `~/.config/tars/profiles/<alias>.env`: per-account tokens, sourced while that profile is
  active (`tars profiles use`); seeded once, never overwritten. Profiles derive paths by
  convention: dir `<projects_dir>/<name>`, key `~/.ssh/id_rsa.<alias>`; add no explicit
  dir or key fields.

## Releasing

Every merge to `main` deploys: `release.yml` bumps the patch tag and runs GoReleaser
(darwin/linux × amd64/arm64, GitHub Release, Homebrew tap); a failed run deletes its release.
Merges touching only the root `README.md`/`CLAUDE.md`, `docs/` or `vhs/` skip it; any other
`.md` is embedded in the binary, so it releases. Minor/major bumps are a manual tag push; see
`docs/releasing.md`.

## Platform notes

macOS-first (Homebrew, `~/Library/Fonts`, iTerm2/Terminal.app plists), Linux-supported with
apt/tarball installs for shared bastions; nothing needs sudo except apt. Neovim is the primary
editor (needs Python3, Node.js, Mason-installed language servers); vim is a minimal fallback.
