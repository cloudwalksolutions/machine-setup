
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"

if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({
    "git",
    "clone",
    "--filter=blob:none",
    "https://github.com/folke/lazy.nvim.git",
    "--branch=stable",
    lazypath
  })
end

local lazy_opts = {
  ui = {
    border = "rounded",
    title = "Plugin Manager",
    title_pos = "center",
  },
}

vim.opt.rtp:prepend(lazypath)

local firenvim_not_active = function()
  return not vim.g.started_by_firenvim
end

local plugin_specs = {

  -- Plenary
  "nvim-lua/plenary.nvim",

  {
    "nvim-tree/nvim-tree.lua",
    dependencies = {
      "nvim-tree/nvim-web-devicons",
    },
    config = function()
      require("config.nvim-tree")
    end,
  },

  -- Sublime Monokai Color Scheme
  {
    "tanvirtin/monokai.nvim",
    config = function()
      -- Set the Monokai color scheme (by default it's 'monokai')
      -- Other variants: monokai_soda, monokai_pro, etc.
      vim.cmd("colorscheme monokai")
    end,
  },

  -- Treesitter
  {
    "nvim-treesitter/nvim-treesitter",
    build = ":TSUpdate",
    config = function()
      require("config.treesitter")
    end,
  },

  -- Completion
  {
    "hrsh7th/nvim-cmp",
    dependencies = {
      "hrsh7th/cmp-nvim-lsp",
      "hrsh7th/cmp-buffer",
      "hrsh7th/cmp-path",
      "hrsh7th/cmp-cmdline",
      "L3MON4D3/LuaSnip",
      "saadparwaiz1/cmp_luasnip",
      "rafamadriz/friendly-snippets"
    },
    config = function()
      require("config.cmp")
    end,
  },

  -- Telescope (fuzzy finder)
  {
    "nvim-telescope/telescope.nvim",
    version = "0.1.x",
    dependencies = { "nvim-lua/plenary.nvim" },
    config = function()
      require("config.telescope")
    end,
  },

  -- Bufferline (buffer management)
  {
    "akinsho/bufferline.nvim",
    version = "*",
    dependencies = {
      "nvim-tree/nvim-web-devicons",
    },
    config = function()
      require("config.bufferline")
    end,
  },

  -- AutoSave
  {
    "pocco81/auto-save.nvim",
    config = function()
      require("auto-save").setup({
        -- optional configuration
        enabled = true, -- start auto-save enabled
        execution_message = {
          message = function() -- message to show on save
            return ("AutoSave: saved at " .. vim.fn.strftime("%H:%M:%S"))
          end,
        },
        events = {"InsertLeave", "TextChanged"},
        conditions = {
          exists = true,
          filename_is_not = {},
          filetype_is_not = {},
          modifiable = true
        },
        -- you can specify write_all_buffers = true to write all
        -- buffers when auto-saving (default is false)
      })
    end,
  },

  -- The missing auto-completion for cmdline!
  {
    "gelguy/wilder.nvim",
    build = ":UpdateRemotePlugins",
  },

  -- show and trim trailing whitespaces
  { "jdhao/whitespace.nvim", event = "VeryLazy" },

  -- NPM Updates
  {
    "vuki656/package-info.nvim",
    dependencies = { "MunifTanjim/nui.nvim" },
    config = function()
      require('config.package-info')
    end
  },

  -- Highlight URLs inside vim
  { "itchyny/vim-highlighturl", event = "VeryLazy" },

  -- notification plugin
  {
    "rcarriga/nvim-notify",
    event = "VeryLazy",
    config = function()
      require("config.nvim-notify")
    end,
  },

  -- Dotenv
  {
    "ellisonleao/dotenv.nvim",
    config = function()
      require('dotenv').setup({
        enable_on_load = true,
        verbose = false,
      })
    end
  },

  -- Toggle Term
  {
    'akinsho/toggleterm.nvim',
    version = "*",
    config = function()
      require("config.toggleterm")
    end
  },

  -- Status Line
  {
    "nvim-lualine/lualine.nvim",
    event = "VeryLazy",
    cond = firenvim_not_active,
    config = function()
      require("config.statusline")
    end
  },

  -- DAP
  {
    'mfussenegger/nvim-dap-python',
    config = function()
      require('config.dap-python')
    end,
  },

  -- Debugger plugin
  {
    "sakhnik/nvim-gdb",
    enabled = function()
      if vim.g.is_win or vim.g.is_linux then
        return true
      end
      return false
    end,
    build = { "bash install.sh" },
    lazy = true,
  },

  -- DAP Extensions
  {
    'mfussenegger/nvim-dap',
    version = "*",
    event = "VeryLazy",
    dependencies = {
      { "theHamsta/nvim-dap-virtual-text", opts = { virt_text_pos = 'eol' }, },

      "nvim-telescope/telescope-dap.nvim",

      {
        "leoluz/nvim-dap-go",
        module = "dap-go",
        opts = {}
      },

      { 'nvim-treesitter/nvim-treesitter' },

      -- fancy UI for the debugger
      {
        "rcarriga/nvim-dap-ui",
        dependencies = {
          "nvim-neotest/nvim-nio",
        },
        -- stylua: ignore
        keys = {
          { ",di", function() require("dapui").toggle({}) end, desc = "Dap UI" },

          {
            ",de",
            function()
              -- Calling this twice to open and jump into the window.
              require('dapui').eval()
              require('dapui').eval()
            end,
            mode = { "n", "v" },
            desc = "Evaluate"
          },
        },
        opts = {
          icons = { expanded = "▾", collapsed = "▸" },
          -- floating = { border = 'rounded' },
          layouts = {
            {
              elements = {
                -- { id = 'stacks',      size = 0.30 },
                -- { id = 'breakpoints', size = 0.20 },
                { id = 'scopes',  size = 0.50 },
                { id = 'watches', size = 0.50 },
              },
              position = 'right',
              size = 40,
            },
            {
              elements = {
                { id = "repl",    size = 0.5 },
                { id = "console", size = 0.5 }
              },
              position = "bottom",
              size = 10
            }
          },
          floating = {
            max_height = nil,
            max_width = nil,
            border = "rounded",
            mappings = {
              close = { "q", "<Esc>" },
            },
          },
          render = {
            max_type_length = nil,
          },
        },
        config = function(_, opts)
          local dapui = require("dapui")
          dapui.setup(opts)
        end
      },

      -- mason.nvim integration
      {
        "jay-babu/mason-nvim-dap.nvim",
        dependencies = "mason.nvim",
        cmd = { "DapInstall", "DapUninstall" },
        opts = {
          -- Makes a best effort to setup the various debuggers with
          -- reasonable debug configurations
          automatic_setup = true,
          -- You can provide additional configuration to the handlers,
          -- see mason-nvim-dap README for more information
          handlers = {},
          -- You'll need to check that you have the required things installed
          -- online, please don't ask me how to install them :)
          ensure_installed = {
            -- Update this to ensure that you have the debuggers for the langs you want
          },
        },
      },

      -- JS/TS debugging.
      {
        "mxsdev/nvim-dap-vscode-js",
        dependencies = {
          {
            'microsoft/vscode-js-debug',
            build = 'npm i && npm run compile vsDebugServerBundle && rm -rf out && mv -f dist out',
          },
        },

      },

      -- Lua adapter.
      {
        "jbyuki/one-small-step-for-vimkind",
        -- stylua: ignore
        keys = {
          { ",daL", function() require("osv").launch({ port = 8086 }) end, desc = "Adapter Lua Server" },
          { ",dal", function() require("osv").run_this() end,              desc = "Adapter Lua" },
        },
        config = function()
          local dap = require("dap")
          dap.adapters.nlua = function(callback, config)
            callback({ type = "server", host = config.host or "127.0.0.1", port = config.port or 8086 })
          end
          dap.configurations.lua = {
            {
              type = "nlua",
              request = "attach",
              name = "Attach to running Neovim instance",
            },
          }
        end,
      },
    },
  },

  -- LazyGit
  {
    "kdheepak/lazygit.nvim",
    cmd = {
        "LazyGit",
        "LazyGitConfig",
        "LazyGitCurrentFile",
        "LazyGitFilter",
        "LazyGitFilterCurrentFile",
    },
    dependencies = {
        "nvim-lua/plenary.nvim",
    },
  },

  -- Search make recipes
  { 'sopa0/telescope-makefile' },

  -- Go Helpers
  {
    "ray-x/go.nvim",
    dependencies = {
      "ray-x/guihua.lua",
      "neovim/nvim-lspconfig",
      "nvim-treesitter/nvim-treesitter",
    },
    config = function()
      require("config.go")
    end,
    event = {"CmdlineEnter"},
    ft = {"go", 'gomod'},
    build = ':lua require("go.install").update_all_sync()' -- if you need to install/update all binaries
  },

  -- Test Coverage
  {
    "andythigpen/nvim-coverage",
    dependencies = { "nvim-lua/plenary.nvim" },
    config = function()
      require("config.coverage")
    end
  },

  -- Easily delete buffers
  { "famiu/bufdelete.nvim" },

  -- fancy start screen
  {
    "nvimdev/dashboard-nvim",
    cond = firenvim_not_active,
    config = function()
      require("config.dashboard-nvim")
    end,
  },

  -- Automatic insertion and deletion of a pair of characters
  { "Raimondi/delimitMate", event = "InsertEnter" },

  -- Manage your yank history
  {
    "gbprod/yanky.nvim",
    cmd = { "YankyRingHistory" },
    config = function()
      require("config.yanky")
    end,
  },

  -- Show git change (change, delete, add) signs in vim sign column
  {
    "lewis6991/gitsigns.nvim",
    config = function()
      require("config.gitsigns")
    end,
  },

  -- K8s
  {
    "hsalem7/nvim-k8s",
    config = function()
      vim.g.vim_k8s_toggle_key_map = ',k'
    end
  },

  -- cucumber-ls
  { "yaegassy/coc-cucumber" },

  -- helm-ls
  { "towolf/vim-helm" },

  -- Python indent (follows the PEP8 style)
  { "Vimjas/vim-python-pep8-indent", ft = { "python" } },

  -- Python-related text object
  { "jeetsukumaran/vim-pythonsense", ft = { "python" } },

  -- Plugin to manipulate character pairs quickly
  { "machakann/vim-sandwich", event = "VeryLazy" },

  -- Illuminate
  { "RRethy/vim-illuminate" },

  -- Markdown
  {
    "toppair/peek.nvim",
    event = { "VeryLazy" },
    build = "deno task --quiet build:fast",
    config = function()
        require("peek").setup()
        vim.api.nvim_create_user_command("PeekOpen", require("peek").open, {})
        vim.api.nvim_create_user_command("PeekClose", require("peek").close, {})
    end,
  },

  -- Toggler
  {
    "nguyenvukhang/nvim-toggler",
    config = function()
      require("config.toggler")
    end
  },

  -- Asynchronous command execution
  { "skywind3000/asyncrun.vim", lazy = true, cmd = { "AsyncRun" } },

  -- Toml
  { "cespare/vim-toml", ft = { "toml" }, branch = "main" },

  -- Finer looking cmdline
  {
    "VonHeikemen/fine-cmdline.nvim",
    dependencies = { "MunifTanjim/nui.nvim" },
    config = function()
      require("config.fine-cmdline")
    end
  },

  -- ChatGPT
  {
    "jackMort/ChatGPT.nvim",
    dependencies = {
      "MunifTanjim/nui.nvim",
      "nvim-lua/plenary.nvim",
      "folke/trouble.nvim",
      "nvim-telescope/telescope.nvim"
    },
    event = "VeryLazy",
    config = function()
      require("config.chatgpt")
    end,
  },

  -- showing keybindings
  {
    "folke/which-key.nvim",
    event = "VeryLazy",
    config = function()
      require("config.which-key")
    end,
  },

  -- Data files
  {
    'vidocqh/data-viewer.nvim',
    opts = {},
    dependencies = {
      "nvim-lua/plenary.nvim",
      "kkharji/sqlite.lua", -- Optional, sqlite support
    }
  },

  -- Github Copilot
  { 'github/copilot.vim' },

  -- LSP
  {
    "neovim/nvim-lspconfig",
    config = function()
      require("lspconfig")
      require("config.lsp")
    end,
  },

  -- LSPKind for showing icons in completion
  {
    "onsails/lspkind-nvim",
    event = "VeryLazy",
    config = function()
      require("config.lspkind")
    end,
  },

  -- Mason for managing external editor tooling
  {
    "williamboman/mason.nvim",
    config = function()
      require("config.mason")
    end,
  },

  -- Mason LSPconfig to bridge mason and lspconfig
  {
    "williamboman/mason-lspconfig.nvim",
    dependencies = { "williamboman/mason.nvim" },
  },

  -- Additional Plugins (e.g., rust-tools)
  {
    "simrat39/rust-tools.nvim",
    ft = "rust",
    config = function()
      require("rust-tools").setup({
        server = {
          on_attach = require("config.lsp").on_attach,
          capabilities = require("config.lsp").capabilities,
          settings = {
            ["rust-analyzer"] = {
              assist = {
                importMergeBehavior = "last",
                importPrefix = "by_self",
              },
              cargo = {
                loadOutDirsFromCheck = true,
              },
              procMacro = {
                enable = true,
              },
            },
          },
        },
      })
    end,
  },

  --- Typescript lsp tools
  {
    "pmizio/typescript-tools.nvim",
    dependencies = { "nvim-lua/plenary.nvim", "neovim/nvim-lspconfig" },
    config = function()
      require("config.tstools")
    end,
    opts = {},
  },

  -- helm.vim for Helm syntax
  {
    "towolf/vim-helm",
    ft = { "helm", "yaml" },
  },

  -- Handy unix command inside Vim (Rename, Move etc.)
  { "tpope/vim-eunuch", cmd = { "Rename", "Delete" } },

  -- Grammar checker
  {
    "rhysd/vim-grammarous",
    enabled = function()
      if vim.g.is_mac then
        return true
      end
      return false
    end,
    ft = { "markdown" },
  },


  -- Plugin to manipulate character pairs quickly
  { "machakann/vim-sandwich", event = "VeryLazy" },

  -- Show logical map of readme contents
  {
    "Zeioth/markmap.nvim",
    build = "yarn global add markmap-cli",
    cmd = { "MarkmapOpen", "MarkmapSave", "MarkmapWatch", "MarkmapWatchStop" },
    opts = {
      html_output = "/tmp/markmap.html", -- (default) Setting a empty string "" here means: [Current buffer path].html
      hide_toolbar = false, -- (default)
      grace_period = 3600000 -- (default) Stops markmap watch after 60 minutes. Set it to 0 to disable the grace_period.
    },
    config = function(_, opts) require("markmap").setup(opts) end
  },

  -- Evaluate commands in markdown
  {
    "jubnzv/mdeval.nvim",
    config = function()
      require("config.mdeval")
    end
  },

  -- CSV Files
  { "mechatroner/rainbow_csv" },

  -- AI/Claude Code
    {
    "greggh/claude-code.nvim",
    dependencies = {
      "nvim-lua/plenary.nvim",
    },
    config = function()
      require("config.claude-code")
    end
  },

  -- {
  --   "coder/claudecode.nvim",
  --   dependencies = { "folke/snacks.nvim" },
  --   config = true,
  --   keys = {
  --     { "<leader>c", nil, desc = "AI/Claude Code" },
  --     { "<leader>cc", "<cmd>ClaudeCode<cr>", desc = "Toggle Claude" },
  --     { "<leader>cf", "<cmd>ClaudeCodeFocus<cr>", desc = "Focus Claude" },
  --     { "<leader>cr", "<cmd>ClaudeCode --resume<cr>", desc = "Resume Claude" },
  --     { "<leader>cC", "<cmd>ClaudeCode --continue<cr>", desc = "Continue Claude" },
  --     { "<leader>cb", "<cmd>ClaudeCodeAdd %<cr>", desc = "Add current buffer" },
  --     { "<leader>cs", "<cmd>ClaudeCodeSend<cr>", mode = "v", desc = "Send to Claude" },
  --     {
  --       "<leader>as",
  --       "<cmd>ClaudeCodeTreeAdd<cr>",
  --       desc = "Add file",
  --       ft = { "NvimTree", "neo-tree", "oil" },
  --     },
  --     -- Diff management
  --     { "<leader>ca", "<cmd>ClaudeCodeDiffAccept<cr>", desc = "Accept diff" },
  --     { "<leader>cd", "<cmd>ClaudeCodeDiffDeny<cr>", desc = "Deny diff" },
  --   },
  -- },

  -- Cursor
  {
    "yetone/avante.nvim",
    event = "VeryLazy",
    build = "make",
    opts = {
      provider = "claude",

      providers = {
        openai = {
          endpoint = "https://api.openai.com/v1",
          model = "gpt-3", -- your desired model (or use gpt-4o, etc.)
          extra_request_body = {
            timeout = 30000, -- Timeout in milliseconds, increase this for reasoning models
            temperature = 0.75,
            max_completion_tokens = 8192, -- Increase this to include reasoning tokens (for reasoning models)
            reasoning_effort = "high", -- low|medium|high, only used for reasoning models
          },
        },
        claude = {
          endpoint = "https://api.anthropic.com",
          model = "claude-opus-4-5-20251101",
          extra_request_body = {
            temperature = 0,
            max_tokens = 8192,
          },
        },
      }
    },
    dependencies = {
      "nvim-treesitter/nvim-treesitter",
      "stevearc/dressing.nvim",
      "nvim-lua/plenary.nvim",
      "MunifTanjim/nui.nvim",
      "nvim-tree/nvim-web-devicons",
      "zbirenbaum/copilot.lua",
      {
        "HakonHarnes/img-clip.nvim",
        event = "VeryLazy",
        opts = {
          default = {
            embed_image_as_base64 = false,
            prompt_for_file_name = false,
            drag_and_drop = {
              insert_mode = true,
            },
            use_absolute_path = true,
          },
        },
      },
      {
        "MeanderingProgrammer/render-markdown.nvim",
        opts = {
          file_types = { "markdown", "Avante" },
        },
        ft = { "markdown", "Avante" },
      },
    },
  }

}

require("lazy").setup(plugin_specs, lazy_opts)
