# Machine Setup

One-command macOS machine setup for provisioning a fully configured development environment.

## Quick Start

```bash
# Fresh machine setup
make init

# Pull latest configs
make pull

# Push local changes
make push
```

## Overview

This repository manages dotfiles and development environment configurations with:

- **One-command setup**: Go from fresh MacBook to production-ready in minutes
- **Automatic backups**: Semantic versioning (v1, v2, v3...) before any changes
- **Modular design**: Pull/push individual components or all at once
- **Easy to customize**: Personal overrides with `.zshrc_secret` and `.zshrc_funcs`

## What's Included

- **Neovim**: IDE-quality configuration with LSP, debugging, AI integrations
- **Zsh**: oh-my-zsh with Powerlevel10k theme, extensive aliases
- **Byobu/tmux**: Terminal multiplexer with custom status bar
- **Vim**: Fallback configuration with Monokai theme
- **Fonts**: Hack Nerd Fonts for proper icon display

## Usage

### Main Commands

```bash
make help        # Show all available commands
make init        # Fresh machine setup (install deps + pull configs)
make pull        # Pull all configs from repo → local
make push        # Push all configs from local → repo
```

### Component-Specific Commands

```bash
make pull-nvim   # Pull only Neovim config
make pull-zsh    # Pull only Zsh config
make pull-byobu  # Pull only Byobu config
make pull-vim    # Pull only Vim config
make pull-fonts  # Install fonts only

make push-nvim   # Push only Neovim config
make push-zsh    # Push only Zsh config
make push-byobu  # Push only Byobu config
make push-vim    # Push only Vim config
```

### Backup Management

```bash
make backup        # Create manual backup of all configs
make backups-list  # List all backup versions
```

Backups are versioned semantically (v1, v2, v3...) and stored in `backups/<component>/vN/`.

## Fresh Machine Setup

On a new MacBook:

```bash
# Clone this repository
git clone <repo-url> ~/machine-setup
cd ~/machine-setup

# Run setup (installs brew packages + applies all configs)
make init

# Edit your personal secrets
vim ~/.zshrc_secret

# Restart terminal
```

## Customization

### Personal Files (Git-Ignored)

- `~/.zshrc_secret` - API keys, tokens, private env vars
- `~/.zshrc_funcs` - Personal shell functions (synced if you want)

### Synced Files

All configuration files in this repo are synced bidirectionally:
- Neovim: `~/.config/nvim/`
- Zsh: `~/.zshrc`, `~/.zshrc_aliases`, `~/.zshrc_funcs`, `~/.profile`
- Byobu: `~/.byobu/`
- Vim: `~/.vimrc`, `~/.vim/colors/`

## Development Workflow

1. **Make changes locally**: Edit files in `~/.config/nvim`, `~/.zshrc`, etc.
2. **Test your changes**: Restart terminal, open nvim, verify everything works
3. **Push to repo**: `make push` (automatically commits and pushes)
4. **Sync to other machines**: `git pull && make pull`

## Architecture

```
machine-setup/
├── Makefile                    # User interface
├── scripts/
│   ├── lib/                    # Shared utilities
│   ├── components/             # Component-specific scripts
│   ├── pull.sh                 # Pull orchestrator
│   ├── push.sh                 # Push orchestrator
│   └── init.sh                 # Init orchestrator
├── backups/                    # Versioned backups (git-ignored)
├── nvim/                       # Neovim configuration
├── zsh/                        # Zsh configuration
├── byobu/                      # Byobu configuration
├── vim/                        # Vim configuration
└── fonts/                      # Hack Nerd Fonts
```

## Migration from Old Version

### Migrating from copy.sh

If you were using the old `copy.sh` script:

```bash
# Old way (deprecated)
./copy.sh pull
./copy.sh push

# New way
make pull
make push
```

The old script still works but shows a deprecation warning.

### Migrating File Names

If you have old dot-based naming (`.zshrc.aliases`, `.zshrc.funcs`, `.zshrc.secret`):

```bash
make migrate
```

This will:
- Copy `.zshrc.aliases` → `.zshrc_aliases`
- Copy `.zshrc.funcs` → `.zshrc_funcs`
- Copy `.zshrc.secret` → `.zshrc_secret`
- Optionally remove old files after successful migration

## Requirements

- macOS (tested on macOS 11+)
- Homebrew (installed by init script if needed)
- Git

## See Also

- [CLAUDE.md](CLAUDE.md) - Detailed architecture documentation for Claude Code
- [nvim/CLAUDE.md](nvim/CLAUDE.md) - Neovim-specific documentation
