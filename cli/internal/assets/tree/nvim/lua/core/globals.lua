
_G.map = vim.keymap.set
_G.nmap = vim.api.nvim_set_keymap

_G.augroup = vim.api.nvim_create_augroup
_G.autocmd = vim.api.nvim_create_autocmd
_G.usercmd = vim.api.nvim_create_user_command

_G.opts = function(desc)
  return { desc = desc, noremap = true, silent = true, nowait = true }
end

