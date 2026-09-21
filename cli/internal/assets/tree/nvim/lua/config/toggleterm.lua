require("toggleterm").setup({
  size = 20,
  open_mapping = [[<c-\>]],
  hide_numbers = true,
  shade_terminals = true,
  autochdir = true, -- Terminal changes directory with Neovim
  start_in_insert = true,
  insert_mappings = true,
  persist_size = true,
  persist_mode = true, -- Remember previous terminal mode
  direction = 'float',
  close_on_exit = true, -- Close terminal window on process exit
  shell = vim.o.shell,  -- Use the system's default shell
  float_opts = {
    border = 'single',
  },
  winbar = {
    enabled = false,
    name_formatter = function(term) -- Customize terminal name in winbar
      return term.name
    end
  },
})

