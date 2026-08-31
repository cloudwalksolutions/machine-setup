
-- Cmdline prompt
map({ "n", "v", "x" }, ":", "<cmd>FineCmdline<CR>", opts("open command line prompt"))
map({ "n", "v", "x" }, ";", "<cmd>FineCmdline<CR>", opts("open command line prompt"))

-- JK to enter Normal mode
map({"i", "t"}, "jk", "<Esc>", opts("Enter command mode"))

-- <C-j> to enter Normal mode in terminal
nmap('t', '<C-j>', [[<C-\><C-n>]], opts("Switch to Normal Mode"))

-- Clear search highlighting
map("n", "<leader>h", ":noh<CR>", opts("Clear search highlight"))

-- Saves the file
map("n", "<leader>w", ":w<CR>", opts("Save file"))

-- Delete current buffer
map({"n", "v"}, "<leader>d", ":Bdelete<CR>", opts("delete current buffer"))

-- Quit all opened buffers
map("n", "<leader>Q", "<cmd>qa!<CR>", opts("quit nvim"))

-- Move windows
map("n", "<C-h>", "<C-w>h", opts("move window left"))
map("n", "<C-l>", "<C-w>l", opts("move window right"))

-- Remove trailing whitespace
map("n", "<Space><Space>", "<cmd>StripTrailingWhitespace<CR>", opts("remove trailing space"))

-- Move current line up and down
map("n", "<C-j>", "V:move '>+1<CR>gv-gv<Esc>", opts("move line down"))
map("n", "<C-k>", "V:move '<-2<CR>gv-gv<Esc>", opts("move line up"))

-- Move highlighted lines up and down
map("x", "J", ":move '>+1<CR>gv-gv")
map("x", "K", ":move '<-2<CR>gv-gv")

-- Resize windows
map("n", "<C-i>", ":resize -3<CR>")
map("n", "<C-m>", ":resize +3<CR>")
map("n", "<C-u>", ":vertical resize -3<CR>")
map("n", "<S-u>", ":vertical resize +3<CR>")

-- Do not move cursor when joining lines.
map("n", "J", function()
  vim.cmd([[
      normal! mzJ`z
      delmarks z
    ]])
end, opts("join lines without moving cursor"))

-- Replace visual selection with text in register, but not contaminate the register
map("x", "p", '"_c<Esc>p')

-- Load nvim config
map({ "n", "v" }, "<leader>sv", function()
  vim.cmd([[
      update $MYVIMRC
      source $MYVIMRC
    ]])
  vim.notify("Nvim config successfully reloaded!", vim.log.levels.INFO, { title = "nvim-config" })
end, opts("reload init.lua"))

-- Log access commands
local logs = require('core.logs')

vim.api.nvim_create_user_command('NvimLog', logs.open_nvim_log, { desc = 'Open Neovim messages log' })
vim.api.nvim_create_user_command('LspLog', logs.open_lsp_log, { desc = 'Open LSP log for current buffer' })
vim.api.nvim_create_user_command('MasonLog', logs.open_mason_log, { desc = 'Open Mason installation log' })
vim.api.nvim_create_user_command('PluginErrors', logs.show_plugin_errors, { desc = 'Show plugin loading errors' })
vim.api.nvim_create_user_command('LogInfo', logs.show_log_locations, { desc = 'Show all log file locations' })

