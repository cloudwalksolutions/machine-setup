require("yanky").setup({
  ring = {
    history_length = 50,
    storage = "memory",
  },
  preserve_cursor_position = {
    enabled = false,
  },
})

-- cycle through the yank history, only work after paste
map("n", "[y", "<Plug>(YankyCycleForward)", opts("cycle forward in yank history"))
map("n", "]y", "<Plug>(YankyCycleBackward)", opts("cycle backward in yank history"))
