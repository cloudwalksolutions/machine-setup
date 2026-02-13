
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

-- Set global defaults for all LSP servers
vim.lsp.config('*', {
  capabilities = M.capabilities,
  flags = {
    debounce_text_changes = 150,
  },
})

-- Move on_attach logic to LspAttach autocmd (new nvim 0.11 pattern)
vim.api.nvim_create_autocmd('LspAttach', {
  callback = function(ev)
    local client = vim.lsp.get_client_by_id(ev.data.client_id)
    if client then
      M.on_attach(client, ev.buf)
    end
  end,
})

-- List of LSP servers to enable
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

-- Enable all servers with global defaults
for _, lsp in ipairs(servers) do
  vim.lsp.enable(lsp)
end

-- Configure terraformls with custom settings
vim.lsp.config('terraformls', {
  filetypes = { "terraform", "tf", "hcl", "tfvars" },
  settings = {
    terraformls = {
      experimentalFeatures = {
        validateOnSave = true
      }
    }
  }
})
vim.lsp.enable('terraformls')

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
