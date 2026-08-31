# tars

`tars` is a one-command CLI that provisions a development environment —
dotfiles, dev tools, fonts, and terminal settings — with automatic, versioned
backups of anything it replaces. The opinionated CloudWalk configs are **embedded
in the binary**, so installing `tars` is all you need; a clone of this repo is
only for contributing.

## Requirements

- **macOS** (primary target) or **Linux** (Debian/Ubuntu; dotfiles fully, installs via apt)
- **[Homebrew](https://brew.sh)** on macOS — `tars` installs packages via `brew` but
  does not install Homebrew itself
- **git** (ships with the Xcode Command Line Tools: `xcode-select --install`)

## Getting started (fresh machine)

```bash
# 1. Install the tars CLI — via Homebrew (macOS):
brew install cloudwalksolutions/homebrew-tap/tars

#    …or the native installer (Linux servers/bastions, or anywhere without brew):
curl -fsSL https://raw.githubusercontent.com/cloudwalksolutions/machine-setup/main/install.sh | sh

# 2. Provision the machine
tars setup
```

The native installer detects OS/arch, verifies the release checksum, and installs a
single binary to `~/.local/bin` (override with `TARS_INSTALL_DIR`) — no sudo, no
dependencies beyond curl + tar.

No clone needed: the dotfiles ship inside the binary and are materialized under
`~/.local/share/tars/repo` on first use. If you DO have a clone (contributors),
`tars` prefers it — found by walking up from your current directory (looking for
`cli/go.mod` + `nvim/`) or via an explicit pointer:

```bash
export MACHINE_SETUP_REPO="$HOME/machine-setup"
```

Verify the install any time with `tars --version`.

## Commands

```bash
tars setup     # full bootstrap: pick tools, install packages, apply all configs
tars pull      # apply repo configs to this machine (dotfiles/fonts/terminals) — no installs
tars push      # copy your local config changes back into the repo
tars sessions  # open byobu sessions from a simple config of dirs (alias: s, by)
```

- **`setup`** is the fresh-machine command. In order, it: shows a welcome screen, lets
  you pick which dev tools to install, installs those packages (Homebrew), installs
  oh-my-zsh and Powerlevel10k, then applies all configs.
- **`pull`** only lays down configuration — no package installs, no network — so it's
  safe to run repeatedly (e.g. after `git pull` to sync new config).
- **`push`** captures your local edits back into the repo so you can commit them
  (requires a real clone — the embedded configs are read-only).

`pull` and `push` exit non-zero when any component fails (each failure is listed),
so scripts and config management can detect partial runs.

For unattended/CI runs, set `MACHINE_SETUP_NO_FORM=1` to skip the interactive prompts
(all offered tools are selected).

## Sessions

Stop rebuilding the same byobu windows after every terminal restart: declare them once
in `~/.config/.machine-setup/sessions.yaml` (git-ignored, machine-specific) — each
session is a name plus a list of dirs, one window per dir:

```yaml
sessions:
  - name: cloudwalk
    dirs:
      - ~/projects/machine-setup
      - ~/projects/api
  - name: personal
    dirs:
      - ~/dotfiles
```

```bash
tars sessions              # interactive picker (or: tars s)
tars sessions all          # open every configured session, attach to the first (tars s a)
tars sessions open <name>  # open just one (tars s o cloudwalk)
tars sessions new [name]   # a fresh byobu session, unrelated to the config (tars s n)
tars sessions list         # show what's configured (tars s l)
tars sessions edit         # edit the config in $EDITOR, seeding an example (tars s e)
```

Opening is idempotent: an existing session is attached, never duplicated, so re-running
after a terminal restart just reconnects. Inside byobu it switches sessions instead of
nesting. (This deliberately avoids byobu's `~/.byobu/windows.tmux`, which creates a
duplicate session set on every launch.)

## ⚠️ What this does to your machine

`tars` writes into your home directory. **Before overwriting anything it makes a
versioned backup** under `~/.local/state/tars/backups/<component>/vN/`, so nothing is
lost — but be aware
it replaces these if they already exist:

- `~/.zshrc`, `~/.zshrc_aliases`, `~/.zshrc_funcs`, `~/.zprofile`
- `~/.config/nvim/` (replaced wholesale)
- `~/.byobu/`, `~/.vimrc`, `~/.vim/colors/`

Also note:

- **These are opinionated CloudWalk defaults** — you'll get our Neovim/Zsh/Byobu setup.
- **Fonts install needs `sudo`.** Copying into `/Library/Fonts` prompts for your password.
- **`setup` runs third-party install scripts** (oh-my-zsh, Powerlevel10k) via their
  official `curl | sh` installers.
- Put personal secrets and per-account aliases in `~/.zshrc_secret` (git-ignored) — see
  `zsh/zshrc_secret.template`.

## What's included

- **Neovim** — IDE-quality config with LSP, debugging, AI integrations
- **Zsh** — oh-my-zsh + Powerlevel10k, aliases, functions
- **Byobu/tmux** — custom status bar and keybindings
- **Vim** — fallback config with Monokai
- **Fonts** — Hack Nerd Font (icon glyphs in nvim/terminal)
- **Terminals** — sets iTerm2 + Terminal.app to the Nerd Font

## Customization (git-ignored)

- `~/.zshrc_secret` — API keys, tokens, per-account aliases (template:
  `zsh/zshrc_secret.template`)
- `~/.zshrc_funcs` — personal shell functions

## How backups work

Every overwrite is archived first, semantically versioned under
`~/.local/state/tars/backups/<component>/vN/` (and `<component>-repo/vN/` for `push`).
Backups are per-user, never auto-deleted, and never written into the repo clone —
so a shared or read-only clone works fine. Override the location with
`MACHINE_SETUP_BACKUP_ROOT`. `tars` skips the copy (and the backup) when a file
already matches, so re-running is a no-op when nothing changed.

## Releasing (maintainers)

Releases are cut by GoReleaser on a semver tag:

```bash
git tag v0.1.0
git push origin v0.1.0     # triggers .github/workflows/release.yml
```

This builds darwin/linux (amd64/arm64) archives + checksums, publishes a GitHub
Release, and updates the Homebrew tap. **One-time prerequisites for the Homebrew push:**

1. Create the tap repo `cloudwalksolutions/homebrew-tap` (empty is fine).
2. Add a `HOMEBREW_TAP_TOKEN` Actions secret — a PAT with write access to that tap repo
   (the default `GITHUB_TOKEN` can't push to another repo).

Test the release config locally without tagging (builds into `./dist`):

```bash
HOMEBREW_TAP_TOKEN=x goreleaser release --snapshot --clean
```

## Development

The CLI is Go, in `cli/` (the module lives there). Tests are Ginkgo/Gomega:

```bash
cd cli && go test ./...
```

CI (`.github/workflows/ci.yml`) runs `go vet` + build + `go test -race` on Linux and
macOS, golangci-lint, and a Docker-isolated end-to-end test
(`test/e2e/Dockerfile`) on every PR to `main`.

Config management lives entirely in `tars`. The `Makefile` is for developing the CLI +
configs: `make build`, `make lint`, `make unit` (fast/airgapped), `make integration`
(brew + Neovim tests), `make test` (unit + integration), `make e2e` (Docker), and
`make check` (lint + build + test). Run `make help` for the full list.

> See [CLAUDE.md](CLAUDE.md) and [nvim/CLAUDE.md](nvim/CLAUDE.md) for architecture and
> the testing/TDD workflow.
