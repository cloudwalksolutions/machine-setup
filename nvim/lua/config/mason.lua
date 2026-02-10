
require("mason").setup({
  ui = {
    border = "rounded",
  },
})

require("mason-lspconfig").setup({
  ensure_installed = {
    "lua_ls",
    "pyright",        -- Python
    "gopls",          -- Go
    "rust_analyzer",  -- Rust
    -- "tsserver",       -- TypeScript/JavaScript
    "zls",            -- Zig
    "bashls",         -- Bash
    "sqls",           -- SQL
    "yamlls",         -- YAML
    "helm_ls",        -- Helm
    "clangd",         -- C/C++
  },
  automatic_installation = true,
})
