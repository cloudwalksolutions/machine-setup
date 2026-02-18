
require("mason").setup({
  ui = {
    border = "rounded",
  },
})

local ensure_lsp = {
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
}

-- Only auto-install clangd if clang toolchain is available on the system
if vim.fn.executable("clang") == 1 or vim.fn.executable("clangd") == 1 then
  table.insert(ensure_lsp, "clangd")
end

require("mason-lspconfig").setup({
  ensure_installed = ensure_lsp,
  automatic_installation = true,
})

-- Ensure DAP tools are installed
local ok, mason_registry = pcall(require, "mason-registry")
if ok then
  local ensure_dap = { "debugpy" }
  for _, tool in ipairs(ensure_dap) do
    local pkg_ok, pkg = pcall(mason_registry.get_package, tool)
    if pkg_ok and not pkg:is_installed() then
      pkg:install()
    end
  end
end
