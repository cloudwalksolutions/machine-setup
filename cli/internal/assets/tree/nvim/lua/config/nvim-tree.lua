local map = vim.keymap.set
local nvim_tree = require("nvim-tree")
local api = require("nvim-tree.api")

local function my_on_attach(bufnr)
  local function n_opts(desc)
    return { desc = "nvim-tree: " .. desc, buffer = bufnr, noremap = true, silent = true, nowait = true }
  end

  api.config.mappings.default_on_attach(bufnr)

  map('n', "<leader>e", api.tree.toggle, n_opts('Toggle Explorer'))
  map('n', '<C-t>', api.tree.change_root_to_node, n_opts('Up'))
  map('n', 'h', api.node.navigate.parent_close, n_opts('Close Node'))
  map('n', '?', api.tree.toggle_help, n_opts('Help'))
end

nvim_tree.setup {
  auto_reload_on_write = true,
  disable_netrw = false,
  hijack_netrw = true,
  hijack_cursor = false,
  hijack_unnamed_buffer_when_opening = false,
  open_on_tab = false,
  sort_by = "name",
  update_cwd = false,
  view = {
    width = 30,
    side = "left",
    preserve_window_proportions = false,
    number = false,
    relativenumber = false,
    signcolumn = "yes",
    adaptive_size = true,
  },
  renderer = {
    indent_markers = {
      enable = false,
      icons = {
        corner = "└ ",
        edge = "│ ",
        none = "  ",
      },
    },
    icons = {
      webdev_colors = true,
    },
  },
  hijack_directories = {
    enable = true,
    auto_open = true,
  },
  update_focused_file = {
    enable = false,
    update_cwd = false,
    ignore_list = {},
  },
  system_open = {
    cmd = "",
    args = {},
  },
  diagnostics = {
    enable = false,
    show_on_dirs = false,
    icons = {
      hint = "",
      info = "",
      warning = "",
      error = "",
    },
  },
  filters = {
    dotfiles = false,
    custom = {},
    git_ignored = false,  -- Show git ignored files
    exclude = {
      "node_modules",
      "__pycache__",
      ".idea",
      ".vscode",
      "cover.out"
    },
  },
  git = {
    enable = true,
    ignore = false,
    timeout = 400,
  },
  actions = {
    use_system_clipboard = true,
    change_dir = {
      enable = true,
      global = false,
      restrict_above_cwd = false,
    },
    open_file = {
      quit_on_open = false,
      resize_window = false,
      window_picker = {
        enable = true,
        chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
        exclude = {
          filetype = { "notify", "qf", "diff", "fugitive", "fugitiveblame" },
          buftype = { "nofile", "terminal", "help" },
        },
      },
    },
  },
  trash = {
    cmd = "trash",
    require_confirm = true,
  },
  log = {
    enable = false,
    truncate = false,
    types = {
      all = false,
      config = false,
      copy_paste = false,
      diagnostics = false,
      git = false,
      profile = false,
    },
  },
}

vim.keymap.set("n", "<leader>e", api.tree.toggle, opts("toggle nvim-tree"))

require("nvim-tree").setup {
  on_attach = my_on_attach,
}


