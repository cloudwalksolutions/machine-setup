# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This is a one-command macOS machine setup repository that provisions a fully configured development environment. The goal is to go from a fresh MacBook to a production-ready development machine with minimal manual intervention.

## Core Philosophy

**Single Command Setup**: Run `./copy.sh init` on a new machine and get:
- Package manager and essential tools (starship, byobu, neovim)
- IDE-quality Neovim configuration with LSP, debugging, AI integrations
- Zsh with oh-my-zsh, plugins, and custom prompt
- Terminal multiplexer (byobu/tmux) with custom status bar
- Development fonts and macOS automation presets
- All environment configurations in their proper locations

**Bidirectional Sync**: The `copy.sh` script manages configuration flow:
- `pull` - Repository → Local machine (apply configs)
- `push` - Local machine → Repository (save changes)
- This enables configuration development in situ with easy version control

**Personalization First**: Configurations are designed to be overridden:
- Zsh sources optional files: `~/.zshrc.secret`, `~/.zshrc.funcs`, `.env`, `./venv/bin/activate`
- Neovim plugin system allows easy additions/removals
- Git ignores personal/secret files by default

## Architecture

### Component Structure

```
├── copy.sh              # Central management script (init|pull|push)
├── nvim/                # Full IDE-quality Neovim config (see nvim/CLAUDE.md)
├── zsh/                 # Shell configuration and aliases
├── byobu/               # Terminal multiplexer config
├── vim/                 # Fallback vim config
├── fonts/               # Hack Nerd Fonts
└── walker.bttpreset     # BetterTouchTool macOS automation
```

### Key Configuration Locations

**Neovim**: Modular Lua architecture with core system (`lua/core/`) and plugin configs (`lua/config/`). See `nvim/CLAUDE.md` for complete documentation.

**Zsh**: Main config in `zsh/zshrc`, aliases in `zsh/zshrc.aliases`. Uses oh-my-zsh with plugins for git, kubectl, fzf, etc.

**Byobu**: Custom status bar and keybindings, including git integration via `byobu/bin/1_git`.

## Development Workflow

### Testing Changes with Claude Code

**TODO: Define formal testing strategy**

Current approach for validating changes:
1. Make modifications to configs in this repository
2. Run `./copy.sh pull` to apply to local environment
3. Test functionality (open nvim, source zshrc, check byobu)
4. If working, run `./copy.sh push` to commit back to repo

**Future considerations:**
- Automated validation scripts for config syntax
- Docker-based testing environment for isolated testing
- Dry-run mode for `copy.sh` operations
- Validation of required dependencies before applying configs

### Making Configuration Changes

1. **Direct editing**: Edit files in home directory (`~/.config/nvim`, `~/.zshrc`) for iterative development
2. **Sync back**: Run `./copy.sh push` when satisfied (automatically commits and pushes)
3. **Pull updates**: Run `./copy.sh pull` or `git pull && ./copy.sh pull` to apply latest from repo

### Adding New Components

When adding new tools or configurations:
1. Add installation command to `copy.sh init` section
2. Add file copy operations to both `pull()` and `push()` functions
3. Update `.gitignore` to exclude generated/personal files
4. Test on clean environment if possible

## Customization and Overrides

### Personal Overrides (Git-Ignored)

The configuration system automatically sources these files if they exist:
- `~/.zshrc.secret` - Private environment variables, API keys
- `~/.zshrc.funcs` - Personal shell functions
- Local `.env` files in project directories
- Project-specific virtual environments

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

- **macOS-specific**: Configurations assume macOS (Homebrew, paths, BetterTouchTool)
- **Destructive operations**: `copy.sh pull` overwrites local configs; commit local changes first
- **Auto-push behavior**: `copy.sh push` automatically commits and pushes; review `git status` at confirmation prompt
- **Neovim as primary editor**: Main development focus is Neovim; vim config is minimal fallback
- **External dependencies**: Neovim config auto-installs LSPs via Mason, but some plugins may require Node.js, Python3, etc.
