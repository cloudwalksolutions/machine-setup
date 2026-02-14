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

### which-key Configuration Pattern

**CRITICAL:** which-key v3 syntax requires `wk.register()` calls OUTSIDE `wk.setup()`:

```lua
-- CORRECT:
wk.setup {
  triggers = {" ", ","},
  -- ... other options
}  -- Close setup here

-- THEN register keybindings
wk.register({ ... }, { prefix = " " })
wk.register({ ... }, { prefix = "," })

-- INCORRECT (will cause syntax error):
wk.setup {
  triggers = {","},
  wk.register({ ... })  -- ❌ INSIDE setup - BREAKS!
}
```

**Dual Registration Pattern:**
- Two separate `wk.register()` calls for two prefixes
- Space prefix for core native (LSP, debug, text editing)
- Comma prefix for plugin commands
- Both prefixes registered in triggers array

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

## Dual-Prefix Keybinding System

**IMPORTANT:** This configuration uses TWO separate leader keys with distinct purposes:
- **`<Space>`** (leader): Core native operations - text editing, LSP, file ops, debug
- **`,`** (comma, localleader): Plugin commands only

Both are registered in which-key for discoverability. Press the prefix key to see available commands.

### Leader Key Configuration
- `vim.g.mapleader = " "` (Space) - defined in `lua/core/settings.lua`
- `vim.g.maplocalleader = ","` (Comma) - defined in `lua/core/settings.lua`
- which-key triggers: `{" ", ","}` - defined in `lua/config/which-key.lua`

### Space Prefix - Core Native (`<Space>` + key)

**LSP (`<Space>l`)** - Language Server Protocol commands:
- `<Space>l` → LSP menu (which-key popup)
- `<Space>lr` → Rename symbol
- `<Space>la` → Code action
- `<Space>lk` → Signature help
- `<Space>lo` → Open diagnostics float
- `<Space>ll` → Location list
- `<Space>lD` → Type definition
- `<Space>lf` → Format code
- `<Space>ld` → Go to definition
- `<Space>li` → Go to implementation
- `<Space>lR` → References
- `<Space>lw` → Workspace submenu (add/remove/list folders)

**Debug (`<Space>d`)** - Debugging commands:
- `<Space>d` → Debug menu
- `<Space>db` → Toggle breakpoint
- `<Space>dB` → Conditional breakpoint
- `<Space>dc` → Continue (Go debugger)
- `<Space>dt` → Debug test (Go)
- `<Space>df` → Debug file (Go)
- `<Space>dp` → Debug selection (Python)

**Text Editing & File Operations:**
- `<Space>t` → Toggle (true/false, yes/no, etc. via nvim-toggler)
- `<Space>e` → Toggle file tree (nvim-tree)
- `<Space>w` → Save file
- `<Space>h` → Clear search highlight
- `<Space>Q` → Quit all buffers
- `<Space>sv` → Reload Neovim configuration

### Comma Prefix - Plugin Commands (`,` + key)

**AI & Code Generation:**
- `,c` → ChatGPT menu (add tests, docstring, fix bugs, optimize, explain, etc.)

**Testing:**
- `,C` → Cypress menu (E2E, spec, component tests)
- `,P` → Playwright menu (E2E, UI mode, debug, codegen)

**Search & Navigation:**
- `,f` → Telescope/Search menu (commands, diagnostics, files, registers, etc.)

**Language-Specific Tools:**
- `,g` → Go tools menu (format, test, coverage, lint, run - NOT debug, debug is on Space)

**Development Tools:**
- `,G` → LazyGit
- `,k` → k9s (Kubernetes)
- `,m` → Make (Telescope make recipes)
- `,t` → Terminal menu (float, horizontal, tab, vertical)
- `,z` → Lazy plugin manager (check, clean, install, update, etc.)

**Content Tools:**
- `,M` → Markdown menu (eval, mindmap, preview, tables)
- `,n` → Node/Package menu (npm/yarn dependency management)
- `,D` → DataViewer

### Non-Leader Bindings (Direct)

**Core Navigation:**
- `<C-f>` - Find files (Telescope)
- `<C-Space>` - Live grep (Telescope)
- `<C-b>` - Buffer picker (Telescope)
- `jk` - Enter normal mode from insert/terminal
- `<C-h>/<C-l>` - Move between windows
- `<C-j>` - Exit terminal mode

**LSP (Direct - also available in `<Space>l` menu):**
- `K` - Hover documentation
- `gd` - Go to definition
- `gr` - Go to references
- `gi` - Go to implementation
- `[d` - Previous diagnostic
- `]d` - Next diagnostic

**Buffer Navigation:**
- `<S-h>` - Previous buffer
- `<S-l>` - Next buffer

## Development Workflows

### Testing & Coverage
- Coverage files supported: `cover.out` (Go)
- Load coverage: `require("core.utils").load_coverage()`
- Visual coverage indicators in gutter

### Debugging
- DAP configuration for Python, Go, JavaScript/TypeScript, Lua
- All debug commands consolidated under `<Space>d` prefix
- Debug commands are on SPACE prefix (core native), NOT comma prefix (plugins)
- Key bindings:
  - `<Space>d` - Debug menu (which-key popup)
  - `<Space>db` - Toggle breakpoint
  - `<Space>dB` - Conditional breakpoint
  - `<Space>dc` - Continue (Go)
  - `<Space>dt` - Debug test (Go)
  - `<Space>df` - Debug file (Go)
  - `<Space>dp` - Debug selection (Python)

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

### Testing System
- Smoke tests in `nvim/tests/smoke_test.lua`
- Run via `make test-nvim` from machine-setup root
- Tests validate:
  - Config loads without Lua errors
  - Critical plugins exist (lazy, lspconfig, cmp, telescope, etc.)
  - LSP configuration is valid
  - Health check module exists
  - Log utilities exist
  - Plugin configurations are valid
- **ALWAYS run tests after configuration changes**

### Health Checks
- Custom health check: `:checkhealth nvim_config`
- Validates external dependencies, Mason tools, LSP servers, API keys, config errors
- Located in `nvim/lua/health/nvim_config.lua`

### Log Access
- `:NvimLog` - Open nvim messages log
- `:LspLog` - Open LSP log for current buffer
- `:MasonLog` - Open Mason installation log
- `:PluginErrors` - Show lazy.nvim plugin errors
- Located in `nvim/lua/core/logs.lua`

## Best Practices

### When Modifying Keybindings
1. Understand the dual-prefix system: Space for core native, Comma for plugins
2. LSP commands go on Space prefix (`<Space>l`)
3. Debug commands go on Space prefix (`<Space>d`)
4. Plugin commands go on Comma prefix (`,`)
5. Update which-key registration in `lua/config/which-key.lua`
6. Keep direct mappings for frequently used commands (speed) AND which-key registration (discoverability)
7. Always run `make test-nvim` after changes

### When Adding New Plugins
1. Add to `lua/core/plugins.lua` with appropriate lazy loading
2. Create config file in `lua/config/[plugin-name].lua`
3. If plugin has commands, add to which-key under comma prefix
4. If plugin enhances core functionality (like LSP), consider space prefix
5. Test thoroughly with `make test-nvim`

### When Fixing Issues
1. Check logs first: `:NvimLog`, `:LspLog`, `:PluginErrors`
2. Run health check: `:checkhealth nvim_config`
3. Ensure which-key syntax is correct (register outside setup)
4. Verify no keybinding conflicts between prefixes
5. Always test fixes with smoke tests

### Machine Setup Integration
- This nvim config is part of the machine-setup repository
- Use `make pull-nvim` to copy repo config to `~/.config/nvim/`
- Use `make push-nvim` to copy `~/.config/nvim/` to repo
- Backups are semantic versioned (v1, v2, v3, etc.) in `backups/nvim/`
- See root README.md for full machine-setup commands