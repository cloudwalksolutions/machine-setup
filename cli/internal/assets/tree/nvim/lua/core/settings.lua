local utils = require('core.utils')

-- Leader Keys
vim.g.mapleader = " "       -- Space for text editing, LSP, debug (core native)
vim.g.maplocalleader = ","  -- Comma for plugin commands


-- General Options
vim.opt.signcolumn = "yes:2"      -- Show two columns for signs
vim.opt.number = true             -- Enable line numbers
vim.opt.relativenumber = true     -- Relative line numbers
vim.opt.termguicolors = true      -- True color support
vim.opt.wrap = false              -- Disable line wrap
vim.opt.tabstop = 2               -- Number of spaces that a <Tab> counts for
vim.opt.shiftwidth = 2            -- Size of an indentation
vim.opt.expandtab = true          -- Use spaces instead of tabs
vim.opt.smartindent = true        -- Autoindent new lines
vim.opt.hidden = true             -- Switch buffers without saving

vim.opt.clipboard = "unnamedplus"
vim.opt.undofile = true
vim.opt.undodir = vim.fn.stdpath("config") .. "/.undo"

-- Use English as main language
vim.cmd [[language en_US.UTF-8]]

-- Disable loading certain plugins

-- Whether to load netrw by default, see https://github.com/bling/dotvim/issues/4
vim.g.loaded_netrw       = 1
vim.g.loaded_netrwPlugin = 1
vim.g.netrw_liststyle = 3
if vim.g.is_win then
  vim.g.netrw_http_cmd = "curl --ssl-no-revoke -Lo"
end

-- Do not load tohtml.vim
vim.g.loaded_2html_plugin = 1

-- Do not load zipPlugin.vim, gzip.vim and tarPlugin.vim (all these plugins are
-- related to checking files inside compressed files)
vim.g.loaded_zipPlugin = 1
vim.g.loaded_gzip = 1
vim.g.loaded_tarPlugin = 1

-- Do not load the tutor plugin
vim.g.loaded_tutor_mode_plugin = 1

-- Do not use builtin matchit.vim and matchparen.vim since we use vim-matchup
vim.g.loaded_matchit = 1
vim.g.loaded_matchparen = 1

-- Disable sql omni completion, it is broken.
vim.g.loaded_sql_completion = 1

-- Markdown syntax
vim.g.markdown_fenced_languages = {'python', 'cpp', 'rust', 'go', 'javascript'}

------------------------------------------------------------------------
--                          custom variables                          --
------------------------------------------------------------------------
vim.g.is_win = (utils.has("win32") or utils.has("win64")) and true or false
vim.g.is_linux = (utils.has("unix") and (not utils.has("macunix"))) and true or false
vim.g.is_mac  = utils.has("macunix") and true or false

vim.g.logging_level = "info"

------------------------------------------------------------------------
--                         builtin variables                          --
------------------------------------------------------------------------
vim.g.loaded_perl_provider = 0  -- Disable perl provider
vim.g.loaded_ruby_provider = 0  -- Disable ruby provider
vim.g.loaded_node_provider = 0  -- Disable node provider
vim.g.did_install_default_menus = 1  -- do not load menu

if utils.executable('python3') then
  if vim.g.is_win then
    vim.g.python3_host_prog = vim.fn.substitute(vim.fn.exepath("python3"), ".exe$", '', 'g')
  else
    vim.g.python3_host_prog = vim.fn.exepath("python3")
  end
else
  vim.api.nvim_err_writeln("Python3 executable not found! You must install Python3 and set its PATH correctly!")
  return
end
