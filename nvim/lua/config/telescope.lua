-- Key mappings
map("n", "<C-f>", function()
  require("telescope.builtin").find_files()
end, { desc = "[F]ind [F]iles" })

map("n", "<C-Space>", function()
  require("telescope.builtin").live_grep()
end, { desc = "[F]ind by [G]rep" })

map("n", "<C-b>", function()
  require("telescope.builtin").buffers()
end, { desc = "[F]ind [B]uffers" })


require('telescope').setup{
  defaults = {
    -- Default configuration for telescope goes here:
    -- config_key = value,
    mappings = {
      i = {
        -- map actions.which_key to <C-h> (default: <C-/>)
        -- actions.which_key shows the mappings for your picker,
        -- e.g. git_{create, delete, ...}_branch for the git_branches picker
        ["<C-h>"] = "which_key"
      }
    }
  },
  pickers = {
    -- Default configuration for builtin pickers goes here:
    -- picker_name = {
    --   picker_config_key = value,
    --   ...
    -- }
    -- Now the picker_config_key will be applied every time you call this
    -- builtin picker
  },
  extensions = {
    -- Your extension configuration goes here:
    -- extension_name = {
    --   extension_config_key = value,
    -- }
    -- please take a look at the readme of the extension you want to configure
    package_info = {
        -- Optional theme (the extension doesn't set a default theme)
        theme = "ivy",
    },
  }
}

require("telescope").load_extension("package_info")
require('telescope').load_extension('make')

