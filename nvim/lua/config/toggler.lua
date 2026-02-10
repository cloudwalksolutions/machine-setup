
map("n", "<leader>t", require("nvim-toggler").toggle, { desc = "Toggle" })

require('nvim-toggler').setup({
  inverses = {
    ['!=='] = '==='
  },
  remove_default_keybinds = true,
  remove_default_inverses = false,
})
