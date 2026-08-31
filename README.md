# tars

[![ci](https://github.com/cloudwalksolutions/machine-setup/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/cloudwalksolutions/machine-setup/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/cloudwalksolutions/machine-setup)](https://github.com/cloudwalksolutions/machine-setup/releases/latest)
![coverage](https://raw.githubusercontent.com/cloudwalksolutions/machine-setup/badges/.badges/main/coverage.svg)
[![go](https://img.shields.io/github/go-mod/go-version/cloudwalksolutions/machine-setup?filename=cli%2Fgo.mod)](cli/go.mod)
[![license](https://img.shields.io/github/license/cloudwalksolutions/machine-setup)](LICENSE)

Provision a development machine — dotfiles, dev tools, fonts, terminal — from a
single binary, with a versioned backup of everything it replaces.

![tars demo](vhs/demo.gif)

The CloudWalk Neovim/Zsh/Byobu environment ships **embedded in the binary**: install
`tars`, run it, done. No repo clone needed (with a clone, tars prefers it).

## Features

- **Full bootstrap or config-only**: `setup` installs everything; `pull` just lays
  down configs — offline, no sudo for dotfiles, safe to re-run
- **Versioned backups**: anything replaced is archived first under
  `~/.local/state/tars/backups/<component>/vN` — per-user, never auto-deleted
- **Idempotent**: unchanged files are skipped; re-runs are no-ops
- **Byobu session manager**: declare sessions as a name + list of dirs; open them
  with two keystrokes, never duplicated
- **macOS and Linux** — including shared bastion hosts (read-only clones, per-user
  backups, unattended mode)
- **Automation-friendly**: `pull`/`push` exit non-zero on any component failure

## Install

```bash
# Homebrew (macOS)
brew install --cask cloudwalksolutions/homebrew-tap/tars

# Native installer (Linux servers, or anywhere without brew)
curl -fsSL https://raw.githubusercontent.com/cloudwalksolutions/machine-setup/main/install.sh | sh

# From source
git clone https://github.com/cloudwalksolutions/machine-setup.git && cd machine-setup/cli && go build -o tars .
```

The installer verifies the release checksum and puts a single binary in
`~/.local/bin` (override with `TARS_INSTALL_DIR`).

## Commands

| Command | Alias | What it does |
|---|---|---|
| `tars setup` | | Full bootstrap: pick tools, install them (brew / apt+tarball), install oh-my-zsh + Powerlevel10k, apply all configs |
| `tars pull` | | Apply configs only — no installs, no network |
| `tars push` | | Capture local config edits back into a repo clone |
| `tars sessions` | `s`, `by` | Open byobu sessions from a simple config |

`pull` and `push` exit non-zero when any component fails (each is listed);
`setup` tolerates config failures so a partial bootstrap stays recoverable.
Set `MACHINE_SETUP_NO_FORM` (any value) to skip all interactive prompts.

## Sessions

Stop rebuilding the same byobu windows after every terminal restart — declare them
once (`tars s e` seeds the file):

```yaml
# ~/.config/.machine-setup/sessions.yaml
sessions:
  - name: cloudwalk        # no '.' or ':' in names; at least one dir
    dirs:
      - ~/work/api         # one window per dir
      - ~/work/infra
```

![tars sessions demo](vhs/sessions.gif)

| | |
|---|---|
| `tars s` | interactive picker |
| `tars s a` | open every session, attach to the first |
| `tars s o <name>` | open one (create-or-attach, idempotent) |
| `tars s n [name]` | fresh session rooted at `~` (a name is required inside tmux) |
| `tars s l` / `tars s e` | list / edit the config |

Inside byobu it switches sessions instead of nesting.

## What it touches

Everything below is archived to `~/.local/state/tars/backups/<component>/vN`
(override the root with `MACHINE_SETUP_BACKUP_ROOT`) before being replaced:

- `~/.zshrc`, `~/.zshrc_aliases`, and the login profile (`~/.zprofile` on macOS,
  `~/.profile` on Linux)
- `~/.config/nvim/` (replaced wholesale), `~/.vimrc`, `~/.vim/colors/`
- `~/.byobu/` configs and status scripts
- Fonts: Hack Nerd Font → `/Library/Fonts` (macOS, needs sudo) or
  `~/.local/share/fonts` (Linux, no sudo)
- macOS only: sets the iTerm2 + Terminal.app font/profile

Never overwritten: `~/.zshrc_secret` (seeded from a template when absent — put
API keys and per-account aliases there) and `~/.zshrc_funcs` (yours entirely).
Machine-specific bits belong in git-ignored `*_local` files the configs source.

## Environment variables

| Variable | Effect |
|---|---|
| `MACHINE_SETUP_REPO` | Use this clone as the config source (beats discovery) |
| `MACHINE_SETUP_NO_FORM` | Skip all TUIs (`setup` selects everything; the sessions picker takes the first entry) |
| `MACHINE_SETUP_BACKUP_ROOT` | Backup location (default `~/.local/state/tars/backups`) |
| `MACHINE_SETUP_CONFIG_PATH` | Config file (default `~/.config/.machine-setup/config.yaml`) |
| `MACHINE_SETUP_SESSIONS_PATH` | Sessions file (default `~/.config/.machine-setup/sessions.yaml`) |
| `TARS_INSTALL_DIR` | Where `install.sh` puts the binary |

## Development

The CLI is Go (module in `cli/`), tested with Ginkgo/Gomega and strict TDD —
see [CLAUDE.md](CLAUDE.md). Useful targets: `make check` (lint + build + tests),
`make unit`, `make e2e` (Docker: three users, a root-owned read-only clone, and a
no-clone install — runs in CI on amd64 + arm64), `make demos` (re-record the GIFs).
After editing any dotfile under `nvim/ zsh/ byobu/ vim/ fonts/ terminal/`, run
`make sync-assets` — a drift guard fails CI otherwise. Releasing:
[docs/releasing.md](docs/releasing.md).

## Known limitations

- The Linux package list is thinner than macOS's (no eza/lazygit/k9s/terraform yet)
- Neovim on Linux installs a pinned upstream tarball (v0.11.6) without checksum pinning
- `tars push` writes into the clone it finds — keep shared clones read-only

## License

[Apache-2.0](LICENSE)
