# tars

`tars` is a one-command CLI that provisions a macOS development environment —
dotfiles, dev tools, fonts, and terminal settings — with automatic, versioned
backups of anything it replaces. It applies the opinionated CloudWalk configs kept in
**this repository**, so you run it from a clone of the repo.

## Requirements

- **macOS** (primary target; a Linux/apt path exists but is partial)
- **[Homebrew](https://brew.sh)** — `tars` installs packages via `brew` but does not
  install Homebrew itself
- **git** (ships with the Xcode Command Line Tools: `xcode-select --install`)

## Getting started (fresh machine)

```bash
# 1. Clone this repo (it holds both the CLI and the configs tars applies)
git clone https://github.com/cloudwalksolutions/machine-setup.git ~/machine-setup
cd ~/machine-setup

# 2. Install the tars CLI
brew install cloudwalksolutions/homebrew-tap/tars
#   (or build from source: cd cli && go build -o tars . && sudo mv tars /usr/local/bin/)

# 3. Provision the machine — run from inside the repo
tars setup
```

`tars` finds the repo by walking up from your current directory (looking for this
repo's `cli/go.mod` + `nvim/`). To run it from anywhere, point it at your clone:

```bash
export MACHINE_SETUP_REPO="$HOME/machine-setup"
```

Verify the install any time with `tars --version`.

## Commands

```bash
tars setup     # full bootstrap: pick tools, install packages, apply all configs
tars pull      # apply repo configs to this machine (dotfiles/fonts/terminals) — no installs
tars push      # copy your local config changes back into the repo
```

- **`setup`** is the fresh-machine command. In order, it: shows a welcome screen, lets
  you pick which dev tools to install, installs those packages (Homebrew), installs
  oh-my-zsh and Powerlevel10k, then applies all configs.
- **`pull`** only lays down configuration — no package installs, no network — so it's
  safe to run repeatedly (e.g. after `git pull` to sync new config).
- **`push`** captures your local edits back into the repo so you can commit them.

For unattended/CI runs, set `MACHINE_SETUP_NO_FORM=1` to skip the interactive prompts
(all offered tools are selected).

## ⚠️ What this does to your machine

`tars` writes into your home directory. **Before overwriting anything it makes a
versioned backup** under `backups/<component>/vN/`, so nothing is lost — but be aware
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
`backups/<component>/vN/` (and `backups/<component>-repo/vN/` for `push`). Backups are
git-ignored and never auto-deleted. `tars` skips the copy (and the backup) when a file
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
