local fineline = require('fine-cmdline')
local fn = fineline.fn

local imap = function(key, command, buffer)
  vim.keymap.set('i', key, command, {buffer = buffer})
end

local set_keymaps = function(buffer)
  imap('<Esc>', '<cmd>stopinsert<CR>', buffer)
  imap('<Up>', fn.up_history, buffer)
  imap('<Down>', fn.down_history, buffer)
  imap('<C-h>', fn.up_search_history, buffer)
  imap('<C-l>', fn.down_search_history, buffer)
  imap('<C-k>', fn.stop_complete_or_previous_item, buffer)
  imap('<C-j>', fn.complete_or_next_item, buffer)
end

fineline.setup({
  cmdline = {
    enable_keymaps = true,
    smart_history = true,
    prompt = ': '
  },
  popup = {
    position = {
      row = '95%',
      col = '3%',
    },
    size = {
      width = '40%',
    },
    border = {
      style = 'rounded',
    },
    win_options = {
      winhighlight = 'Normal:Normal,FloatBorder:FloatBorder',
    },
  },
  hooks = {
    before_mount = function(input)
      -- code
    end,
    after_mount = function(input)
      set_keymaps(input.bufnr)
    end,
    set_keymaps = function(imap, feedkeys)
      -- code
    end
  }
})
