# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This is a one-command macOS machine setup repository that provisions a fully configured development environment. The goal is to go from a fresh MacBook to a production-ready development machine with minimal manual intervention.

## Core Philosophy

**Single Command Setup**: Run `make init` on a new machine and get:
- Package manager and essential tools (starship, byobu, neovim)
- IDE-quality Neovim configuration with LSP, debugging, AI integrations
- Zsh with oh-my-zsh, plugins, and custom prompt
- Terminal multiplexer (byobu/tmux) with custom status bar
- Development fonts and macOS automation presets
- All environment configurations in their proper locations

**Bidirectional Sync**: Makefile commands manage configuration flow:
- `make pull` - Repository → Local machine (apply configs with backups)
- `make push` - Local machine → Repository (save changes with git commit)
- Component-specific commands like `make pull-nvim` for granular control
- This enables configuration development in situ with easy version control

**Versioned Backups**: Automatic semantic versioning before any destructive operations:
- Backups stored in `backups/<component>/v1/`, `v2/`, `v3/`...
- Never lose configuration - all changes are backed up
- Use `make backups-list` to see all versions

**Personalization First**: Configurations are designed to be overridden:
- Zsh sources optional files: `~/.zshrc_secret`, `~/.zshrc_funcs`, `.env`, `./venv/bin/activate`
- Template provided for secrets: `zsh/zshrc_secret.template`
- Neovim plugin system allows easy additions/removals
- Git ignores personal/secret files by default

## Architecture

### Component Structure

```
├── Makefile                    # User-facing interface
├── scripts/
│   ├── lib/
│   │   ├── common.sh          # Shared utilities (backup, logging, validation)
│   │   └── config.sh          # Path constants and configuration
│   ├── components/             # Modular component handlers
│   │   ├── nvim.sh            # Neovim pull/push operations
│   │   ├── zsh.sh             # Zsh pull/push operations
│   │   ├── byobu.sh           # Byobu pull/push operations
│   │   ├── vim.sh             # Vim pull/push operations
│   │   └── fonts.sh           # Fonts installation
│   ├── pull.sh                # Orchestrates all component pulls
│   ├── push.sh                # Orchestrates all component pushes
│   └── init.sh                # Fresh machine setup
├── backups/                    # Versioned backups (git-ignored)
├── nvim/                       # Full IDE-quality Neovim config (see nvim/CLAUDE.md)
├── zsh/                        # Shell configuration and aliases
├── byobu/                      # Terminal multiplexer config
├── vim/                        # Fallback vim config
├── fonts/                      # Hack Nerd Fonts
└── walker.bttpreset            # BetterTouchTool macOS automation
```

### Key Configuration Locations

**Neovim**: Modular Lua architecture with core system (`lua/core/`) and plugin configs (`lua/config/`). See `nvim/CLAUDE.md` for complete documentation.

**Zsh**: Main config in `zsh/zshrc`, aliases in `zsh/zshrc_aliases`, functions in `zsh/zshrc_funcs`. Uses oh-my-zsh with plugins for git, kubectl, fzf, etc. Secret template in `zsh/zshrc_secret.template`.

**Byobu**: Custom status bar and keybindings, including git integration via `byobu/bin/1_git`.

## Development Workflow

### Testing Changes with Claude Code

**Current approach** for validating changes:
1. Make modifications to configs in this repository OR edit locally
2. If editing in repo: Run `make pull` to apply to local environment
3. Test functionality (open nvim, source zshrc, check byobu)
4. If working and edited locally: Run `make push` to commit back to repo

**Backup safety**: All pull/push operations automatically create versioned backups (v1, v2, v3...) before making changes. Use `make backups-list` to view all versions.

**Future considerations:**
- Automated validation scripts for config syntax
- Docker-based testing environment for isolated testing
- Dry-run mode for operations
- Validation of required dependencies before applying configs

### Making Configuration Changes

1. **Direct editing**: Edit files in home directory (`~/.config/nvim`, `~/.zshrc`) for iterative development
2. **Sync back**: Run `make push` when satisfied (automatically commits and pushes)
3. **Pull updates**: Run `git pull && make pull` to apply latest from repo
4. **Component-specific**: Use `make pull-nvim`, `make push-zsh`, etc. for granular control

### Adding New Components

When adding new tools or configurations:
1. Create new component script in `scripts/components/<name>.sh` following the standard interface (pull/push functions)
2. Add path mappings to `scripts/lib/config.sh`
3. Add component to orchestration scripts (`scripts/pull.sh`, `scripts/push.sh`)
4. Add Makefile targets: `pull-<name>` and `push-<name>`
5. Add installation commands to `scripts/init.sh` if needed
6. Update `.gitignore` to exclude generated/personal files
7. Test on clean environment if possible

## Customization and Overrides

### Personal Overrides (Git-Ignored)

The configuration system automatically sources these files if they exist:
- `~/.zshrc_secret` - Private environment variables, API keys (git-ignored, template provided)
- `~/.zshrc_funcs` - Personal shell functions (synced if created)
- Local `.env` files in project directories
- Project-specific virtual environments

During `make init`, a template is copied to `~/.zshrc_secret` for easy setup.

### Modifying Shared Configuration

**For personal style preferences:**
- Fork or maintain personal branch for significant deviations
- Use conditional logic in configs (e.g., hostname-based customization)
- Override keybindings in plugin configs rather than modifying core

**For contributions:**
- Keep changes broadly applicable
- Avoid hardcoding personal preferences in shared files
- Document new features or significant changes

## Important Notes

- **Never commit secrets**: The `.gitignore` excludes secret files (`.zshrc_secret`), but be vigilant
- **Backups are versioned**: All operations create semantic version backups (v1, v2, v3...) - never auto-deleted
- **Neovim dependencies**: Neovim configuration requires Python3, Node.js, and various language servers (auto-installed via Mason)
- **macOS specific**: This configuration is tailored for macOS (Homebrew, paths, BetterTouchTool, etc.)
- **Old copy.sh deprecated**: Use `make` commands instead - `copy.sh` shows deprecation warning but still works
- **Neovim as primary editor**: Main development focus is Neovim; vim config is minimal fallback
