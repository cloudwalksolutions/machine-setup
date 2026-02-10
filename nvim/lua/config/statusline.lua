-- lua/core/config/statusline.lua

-- Helper function to map keys
local function map(mode, lhs, rhs, opts)
  local options = { noremap = true, silent = true }
  if opts then
    options = vim.tbl_extend("force", options, opts)
  end
  vim.api.nvim_set_keymap(mode, lhs, rhs, options)
end

-- Bufferline Configuration
local function setup_bufferline()
  require("bufferline").setup {
    options = {
      numbers = "buffer_id",
      close_command = "bdelete! %d", -- Command to close a buffer
      right_mouse_command = nil,      -- Disable right-click closing
      left_mouse_command = "buffer %d", -- Left-click to go to buffer
      middle_mouse_command = nil,     -- Disable middle-click
      indicator = {
        icon = "▎", -- Indicator icon
        style = "icon",
      },
      buffer_close_icon = "",
      modified_icon = "●",
      close_icon = "",
      left_trunc_marker = "",
      right_trunc_marker = "",
      max_name_length = 18,
      max_prefix_length = 15, -- Prefix used when a buffer is de-duplicated
      tab_size = 10,
      diagnostics = false, -- Disable diagnostics in bufferline
      custom_filter = function(bufnr)
        -- Filter out specific filetypes from bufferline
        local exclude_ft = { "qf", "fugitive", "git" }
        local cur_ft = vim.bo[bufnr].filetype
        local should_filter = vim.tbl_contains(exclude_ft, cur_ft)

        if should_filter then
          return false
        end

        return true
      end,
      show_buffer_icons = false, -- Disable buffer icons
      show_buffer_close_icons = true,
      show_close_icon = true,
      show_tab_indicators = true,
      persist_buffer_sort = true, -- Persist custom sorted buffers
      separator_style = "bar",
      enforce_regular_tabs = false,
      always_show_bufferline = true,
      sort_by = "id",
    },
  }
end

-- Lualine Configuration
local function setup_lualine()
  local fn = vim.fn

  -- Function to display spell status
  local function spell()
    if vim.o.spell then
      return "[SPELL]"
    end
    return ""
  end

  -- Function to show IME state for Chinese users
  local function ime_state()
    if vim.g.is_mac then
      local layout = fn.libcall(vim.g.XkbSwitchLib, "Xkb_Switch_getXkbLayout", "")
      local res = fn.match(layout, [[\v(Squirrel\.Rime|SCIM.ITABC)]])
      if res ~= -1 then
        return "[CN]"
      end
    end
    return ""
  end

  -- Function to detect trailing spaces
  local function trailing_space()
    if not vim.o.modifiable then
      return ""
    end

    for i = 1, fn.line("$") do
      local linetext = fn.getline(i)
      local idx = fn.match(linetext, [[\v\s+$]])

      if idx ~= -1 then
        return string.format("[%d]trailing", i)
      end
    end

    return ""
  end

  -- Function to detect mixed indentation
  local function mixed_indent()
    if not vim.o.modifiable then
      return ""
    end

    local space_pat = [[\v^ +]]
    local tab_pat = [[\v^\t+]]
    local space_indent = fn.search(space_pat, "nwc")
    local tab_indent = fn.search(tab_pat, "nwc")
    local mixed = (space_indent > 0 and tab_indent > 0)
    local mixed_same_line
    if not mixed then
      mixed_same_line = fn.search([[\v^(\t+ | +\t)]], "nwc")
      mixed = mixed_same_line > 0
    end
    if not mixed then
      return ""
    end
    if mixed_same_line ~= nil and mixed_same_line > 0 then
      return "MI:" .. mixed_same_line
    end
    local space_indent_cnt = fn.searchcount({ pattern = space_pat, max_count = 1e3 }).total
    local tab_indent_cnt = fn.searchcount({ pattern = tab_pat, max_count = 1e3 }).total
    if space_indent_cnt > tab_indent_cnt then
      return "MI:" .. tab_indent
    else
      return "MI:" .. space_indent
    end
  end

  -- Function to display Git diff
  local diff = function()
    local git_status = vim.b.gitsigns_status_dict
    if git_status == nil then
      return
    end

    local modify_num = git_status.changed
    local remove_num = git_status.removed
    local add_num = git_status.added

    local info = { added = add_num, modified = modify_num, removed = remove_num }
    return info
  end

  -- Function to display Python virtual environment
  local virtual_env = function()
    if vim.bo.filetype ~= 'python' then
      return ""
    end

    local conda_env = os.getenv('CONDA_DEFAULT_ENV')
    local venv_path = os.getenv('VIRTUAL_ENV')

    if venv_path == nil then
      if conda_env == nil then
        return ""
      else
        return string.format("  %s (conda)", conda_env)
      end
    else
      local venv_name = vim.fn.fnamemodify(venv_path, ':t')
      return string.format("  %s (venv)", venv_name)
    end
  end

  -- Function to display package info (requires package-info.nvim)
  local package_info = function()
    return require('package-info').get_status()
  end

  -- Function to detect terminal buffer
  local function is_terminal()
    return vim.bo.buftype == 'terminal'
  end

  -- Function to get terminal name
  local function get_terminal_name()
    if is_terminal() then
      return vim.fn.expand('<afile>:t')
    end
    return ''
  end

  require("lualine").setup {
    options = {
      icons_enabled = true,
      theme = "auto",
      section_separators = "",
      component_separators = "",
      disabled_filetypes = {},
      always_divide_middle = true,
    },
    sections = {
      lualine_a = { "mode" },
      lualine_b = {
        "branch",
        {
          "diff",
          source = diff,
        },
        {
          virtual_env,
          color = { fg = 'black', bg = "#F1CA81" }
        }
      },
      lualine_c = {
        "filename",
        {
          ime_state,
          color = { fg = "black", bg = "#f46868" },
        },
        {
          spell,
          color = { fg = "black", bg = "#a7c080" },
        },
        {
          "diagnostics",
          sources = { "nvim_diagnostic" },
          symbols = {error = '🆇 ', warn = '⚠️ ', info = 'ℹ️ ', hint = ' '},
        },
        -- Terminal Indicator
        {
          function()
            if is_terminal() then
              return "  Terminal"
            end
            return ""
          end,
          cond = is_terminal,
          color = { fg = "white", bg = "#5F5FFF" },
        },
      },
      lualine_x = {
        {
          package_info,
          color = { fg = 'white' },
        },
      },
      lualine_y = {
        "location",
      },
      lualine_z = {
        {
          trailing_space,
          color = "WarningMsg",
        },
        {
          mixed_indent,
          color = "WarningMsg",
        },
      },
    },
    inactive_sections = {
      lualine_a = {},
      lualine_b = {},
      lualine_c = { "filename" },
      lualine_x = { "location" },
      lualine_y = {},
      lualine_z = {},
    },
    tabline = {
      lualine_a = {},
      lualine_b = {},
      lualine_c = {},
      lualine_x = {},
      lualine_y = {},
      lualine_z = {},
      lualine_a = {
        {
          "buffers",
          mode = 1, -- Shows buffer names
          max_length = vim.o.columns * 0.5, -- Limit to 50% of screen
          symbols = {
            modified = "●", -- Icon for modified buffers
            alternate_file = "#",
            directory = "", -- Icon for directory buffers
          },
          buffer_close_icon = "",
          modified_icon = "●",
          close_icon = "",
          left_trunc_marker = "",
          right_trunc_marker = "",
          show_icons = false, -- Disable icons as per your bufferline config
        }
      },
    },
    extensions = { "quickfix", "fugitive", "nvim-tree" },
  }
end

-- Initialize all configurations
local function setup_all()
  setup_bufferline()
  setup_lualine()
end

return {
  setup = setup_all
}
