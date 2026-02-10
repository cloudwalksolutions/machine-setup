
local opts = { noremap=true, silent=true }

map('n', 'K', '<Cmd>lua vim.lsp.buf.hover()<CR>', opts)
map('n', 'gd', '<Cmd>lua vim.lsp.buf.definition()<CR>', opts)
map('n', 'gi', '<Cmd>lua vim.lsp.buf.implementation()<CR>', opts)
map('n', 'gr', '<Cmd>lua vim.lsp.buf.references()<CR>', opts)
map('n', '[d', '<Cmd>lua vim.diagnostic.goto_prev()<CR>', opts)
map('n', ']d', '<Cmd>lua vim.diagnostic.goto_next()<CR>', opts)
map('n', '<leader>r', '<Cmd>lua vim.lsp.buf.rename()<CR>', opts)
map('n', '<leader>k', '<Cmd>lua vim.lsp.buf.signature_help()<CR>', opts)
map('n', '<leader>wa', '<Cmd>lua vim.lsp.buf.add_workspace_folder()<CR>', opts)
map('n', '<leader>wr', '<Cmd>lua vim.lsp.buf.remove_workspace_folder()<CR>', opts)
map('n', '<leader>wl', '<Cmd>lua print(vim.inspect(vim.lsp.buf.list_workspace_folders()))<CR>', opts)
map('n', '<leader>D', '<Cmd>lua vim.lsp.buf.type_definition()<CR>', opts)
map('n', '<leader>a', '<Cmd>lua vim.lsp.buf.code_action()<CR>', opts)
map('n', '<leader>o', '<Cmd>lua vim.diagnostic.open_float()<CR>', opts)
map('n', '<leader>l', '<Cmd>lua vim.diagnostic.setloclist()<CR>', opts)

local M = {}

-- Function to set up keymaps after LSP attaches to the buffer
M.on_attach = function(client, bufnr)
  -- Optional: Highlight symbol under cursor
  if client.server_capabilities.documentSymbolProvider then
    require("illuminate").on_attach(client)
  end
end

-- Function to set up capabilities, especially for auto-completion
M.capabilities = vim.lsp.protocol.make_client_capabilities()

-- Integrate with nvim-cmp for enhanced auto-completion
local cmp_ok, cmp_nvim_lsp = pcall(require, "cmp_nvim_lsp")
if cmp_ok then
  M.capabilities = cmp_nvim_lsp.default_capabilities(M.capabilities)
end

-- List of LSP servers to set up
local servers = {
  "pyright",
  "gopls",
  "rust_analyzer",
  "ts_ls",
  "zls",
  "bashls",
  "sqls",
  "yamlls",
  "helm_ls",
  "clangd",
}

for _, lsp in ipairs(servers) do
  require("lspconfig")[lsp].setup({
    on_attach = M.on_attach,
    capabilities = M.capabilities,
    flags = {
      debounce_text_changes = 150,
    },
    -- Optional: Server-specific settings
    -- settings = {
    --   pyright = { ... },
    --   gopls = { ... },
    --   rust_analyzer = { ... },
    --   -- etc.
    -- },
  })
end

require("lspconfig").terraformls.setup({
  on_attach = M.on_attach,
  capabilities = M.capabilities,
  flags = {
    debounce_text_changes = 150,
  },
  filetypes = { "terraform", "tf", "hcl", "tfvars" },  -- ensure these are recognized
  settings = {
    terraformls = {
      experimentalFeatures = {
        validateOnSave = true
      }
    }
  }
})

-- Define diagnostic signs with icons
local signs = {
  Error = " ",  -- ex: \u{f057}
  Warn  = " ",  -- ex: \u{f071}
  Hint  = " ",  -- ex: \u{f834}
  Info  = " ",  -- ex: \u{f05a}
}

for type, icon in pairs(signs) do
  local hl = "DiagnosticSign" .. type
  vim.fn.sign_define(hl, { text = icon, texthl = hl, numhl = hl })
end

-- Update capabilities for nvim-cmp
local capabilities = require('cmp_nvim_lsp').default_capabilities(vim.lsp.protocol.make_client_capabilities())

return M
