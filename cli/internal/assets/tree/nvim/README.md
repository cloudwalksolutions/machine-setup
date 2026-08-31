# CloudWalk Neovim Configuration

A Neovim configuration designed for the CloudWalk development environment: lightweight
and fast, with a rich out-of-the-box feature set comparable to modern IDEs like the
[JetBrains suite](https://www.jetbrains.com/) and [Cursor](https://cursor.so/), while
preserving classic Vim muscle memory. Installed and kept in sync by
[`tars`](../README.md).

### Features

- Intuitive, classic keybindings that stay customizable
- LSP support for more languages than you probably need
- Syntax highlighting (defaults to Monokai)
- File explorer, fuzzy finder (keywords, files, buffers, …), status line
- Standard + AI autocompletion, snippets
- Debugging (DAP) support
- Terminal, LazyGit, and K9s integration
- Cursor-like AI plugin

### Keybindings

Bindings use a dual-prefix system — `<Space>` for primary actions and `,` for
secondary ones — documented in detail in [CLAUDE.md](CLAUDE.md).

### Development

After editing anything under `nvim/`, run `make test-nvim` (headless smoke tests)
and `make sync-assets` (refreshes the copy embedded in the tars binary — a CI
drift guard fails otherwise).
