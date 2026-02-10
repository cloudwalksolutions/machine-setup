# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a comprehensive Neovim configuration built around the Lazy.nvim plugin manager with a modular architecture:

- **Core System**: `/lua/core/` contains fundamental modules loaded sequentially by `init.lua`
  - `globals.lua` - Global variables and platform detection
  - `settings.lua` - Vim options and basic configuration  
  - `plugins.lua` - Plugin specifications and lazy loading setup
  - `keymaps.lua` - Key mappings and shortcuts
  - `autocmds.lua` - Auto commands and event handlers
  - `highlights.lua` - Syntax highlighting customizations
  - `utils.lua` - Utility functions including coverage loading

- **Plugin Configurations**: `/lua/config/` contains individual plugin setup files
  - Each major plugin has its own configuration module
  - Configurations are loaded on-demand via lazy loading

## Key Development Patterns

### Plugin Management
- Uses Lazy.nvim with sophisticated loading strategies (event-based, filetype-based, command-based)
- Plugin dependencies are well-structured with proper configuration delegation
- Version locking via `lazy-lock.json` for reproducible setups

### Language Support  
- Multi-language LSP setup through Mason and nvim-lspconfig
- Supported languages: Python, Go, Rust, TypeScript, Zig, Bash, SQL, YAML, Helm, C/C++, Terraform
- Debugging support via nvim-dap for multiple languages
- Test coverage integration with visual indicators

### AI Integration
- Multiple AI providers: GitHub Copilot, ChatGPT, Avante (Cursor-like), Claude Code
- AI-assisted development workflows integrated throughout

## Essential Keybindings

### Core Navigation
- `<Space>` - Leader key
- `<C-f>` - Find files (Telescope)
- `<C-Space>` - Grep search (Telescope) 
- `<C-b>` - Buffer picker (Telescope)
- `jk` - Enter normal mode from insert/terminal

### File Operations
- `<leader>w` - Save file
- `<leader>d` - Delete current buffer
- `<leader>Q` - Quit all buffers
- `<leader>sv` - Reload Neovim configuration

### LSP & Development
- `gd` - Go to definition
- `gr` - Go to references  
- `gi` - Go to implementation
- `K` - Show hover documentation
- `<leader>a` - Code actions
- `<leader>h` - Clear search highlight

### Window/Terminal Management
- `<C-h>/<C-l>` - Move between windows
- `<C-j>` - Exit terminal mode
- Terminal integration via ToggleTerm

## Development Workflows

### Testing & Coverage
- Coverage files supported: `cover.out` (Go)
- Load coverage: `require("core.utils").load_coverage()`
- Visual coverage indicators in gutter

### Debugging
- DAP configuration for Python, Go, JavaScript/TypeScript, Lua
- UI integration through nvim-dap-ui
- Key binding: `,di` - Toggle DAP UI

### Git Integration
- LazyGit integration for Git workflows
- Gitsigns for inline Git status
- Various Git commands available through LazyGit interface

## Configuration Management

### Modifying Configuration
- Plugin specs are in `lua/core/plugins.lua`
- Individual plugin configs in `lua/config/[plugin-name].lua` 
- Core settings in `lua/core/settings.lua`
- Keymaps in `lua/core/keymaps.lua`

### Adding New Plugins
1. Add plugin spec to `plugin_specs` table in `lua/core/plugins.lua`
2. Create corresponding config file in `lua/config/` if needed
3. Use appropriate lazy loading strategy (event, ft, cmd, cond)

### Platform Considerations  
- Cross-platform support with platform detection (`vim.g.is_win`, `vim.g.is_linux`, `vim.g.is_mac`)
- Python3 path auto-detection and validation

## Tools and Dependencies

### External Dependencies
- Python3 (required, path auto-detected)
- Git (for plugin management)
- Language servers installed via Mason
- Node.js/npm (for some plugins)
- Various language-specific tooling managed through Mason

### No Traditional Build System
- No Makefiles or shell scripts - relies on plugin-managed tooling
- Mason handles automatic tool installation
- Language-specific build/test commands available through LSP integration