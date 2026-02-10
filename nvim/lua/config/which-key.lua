local wk = require("which-key")

wk.setup {
  plugins = {
    marks = true, -- shows a list of your marks on ' and `
    registers = true, -- shows your registers on " in NORMAL or <C-r> in INSERT mode
    spelling = {
      enabled = true, -- enabling this will show WhichKey when pressing z= to select spelling suggestions
      suggestions = 9, -- how many suggestions should be shown in the list?
    },
    -- the presets plugin, adds help for a bunch of default keybindings in Neovim
    -- No actual key bindings are created
    presets = {
      operators = true, -- adds help for operators like d, y, ... and registers them for motion / text object completion
      motions = true, -- adds help for motions
      text_objects = true, -- help for text objects triggered after entering an operator
      windows = true, -- default bindings on <c-w>
      nav = true, -- misc bindings to work with windows
      z = true, -- bindings for folds, spelling and others prefixed with z
      g = true, -- bindings for prefixed with g
    },
  },
  -- add operators that will trigger motion and text object completion
  -- to enable all native operators, set the preset / operators plugin above
  defer = { gc = "Comments" },
  replace = {
    -- override the label used to display some keys. It doesn't effect WK in any other way.
    -- For example:
    ["<space>"] = "SPACE",
    ["<cr>"] = "ENTER",
    ["<tab>"] = "TAB",
  },
  icons = {
    breadcrumb = "»", -- symbol used in the command line area that shows your active key combo
    separator = "➜", -- symbol used between a key and it's label
    group = "+", -- symbol prepended to a group
  },
  -- win = {
  --   border = "none", -- none, single, double, shadow
  --   position = "bottom", -- bottom, top
  --   margin = { 0, 0, 0, 0 }, -- extra window margin [top, right, bottom, left]
  --   padding = { 1, 0, 1, 0 }, -- extra window padding [top, right, bottom, left]
  -- },
  layout = {
    height = { min = 1, max = 25 }, -- min and max height of the columns
    width = { min = 20, max = 50 }, -- min and max width of the columns
    spacing = 1, -- spacing between columns
    align = "center", -- align columns left, center or right
  },
  show_help = true, -- show help message on the command line when the popup is visible
  triggers = {","},

  wk.register({

    c = {
      name = "ChatGPT",
      a = { "<cmd>ChatGPTRun add_tests<CR>", "Add Tests", mode = { "n", "v" } },
      c = { "<cmd>ChatGPT<CR>", "Chat" },
      d = { "<cmd>ChatGPTRun docstring<CR>", "Docstring", mode = { "n", "v" } },
      e = { "<cmd>ChatGPTEditWithInstruction<CR>", "Edit with instruction", mode = { "n", "v" } },
      f = { "<cmd>ChatGPTRun fix_bugs<CR>", "Fix Bugs", mode = { "n", "v" } },
      g = { "<cmd>ChatGPTRun grammar_correction<CR>", "Grammar Correction", mode = { "n", "v" } },
      k = { "<cmd>ChatGPTRun keywords<CR>", "Keywords", mode = { "n", "v" } },
      l = { "<cmd>ChatGPTRun code_readability_analysis<CR>", "Code Readability Analysis", mode = { "n", "v" } },
      o = { "<cmd>ChatGPTRun optimize_code<CR>", "Optimize Code", mode = { "n", "v" } },
      s = { "<cmd>ChatGPTRun summarize<CR>", "Summarize", mode = { "n", "v" } },
      t = { "<cmd>ChatGPTRun translate<CR>", "Translate", mode = { "n", "v" } },
      r = { "<cmd>ChatGPTRun roxygen_edit<CR>", "Roxygen Edit", mode = { "n", "v" } },
      x = { "<cmd>ChatGPTRun explain_code<CR>", "Explain Code", mode = { "n", "v" } },
    },

    C = {
      name = "Cypress",
      e = { "<cmd>TermExec mode=horizontal cmd='npx cypress run --browser chrome --headless --e2e'<cr>", "Run E2E Tests", mode = { "n", "v" } },
      s = { "<cmd>TermExec mode=horizontal cmd='npx cypress run --spec %'<cr>", "Run Spec", mode = { "n", "v" } },
      u = { "<cmd>TermExec mode=horizontal cmd='npx cypress run --browser chrome --headless --component'<cr>", "Run Component Tests", mode = { "n", "v" } },
    },

    D = { "<cmd>DataViewer<CR>", "Open Data Viewer", mode = { "n", "v" } },

    f = {
      name = "Search",
      c = { "<cmd>Telescope commands<CR>", "Commands", mode = { "n", "v" } },
      d = { "<cmd>Telescope diagnostics<CR>", "Diagnostics", mode = { "n", "v" } },
      f = { "<cmd>Telescope grep_string<CR>", "Grep for Selected", mode = { "v" } },
      m = { "<cmd>Telescope man_pages<CR>", "Man Pages", mode = { "n", "v" } },
      o = { "<cmd>Telescope oldfiles<CR>", "Recent Files", mode = { "n", "v" } },
      r = { "<cmd>Telescope registers<CR>", "Registers", mode = { "n", "v" } },
      s = { "<cmd>Telescope treesitter<CR>", "Code", mode = { "n", "v" } },
      z = { "<cmd>Telescope current_buffer_fuzzy_find<CR>", "Fuzzy in Buffer", mode = { "n", "v" } },
    },

    g = {
      name = "Go",
      g = { "<cmd>GoSave<CR>", "Format and Imports", mode = { "n", "v" } },
      b = { "<cmd>GoBreakToggle<CR>", "Toggle Breakpoint", mode = { "n", "v" } },
      B = { "<cmd>BreakCondition<CR>", "Conditional Breakpoint", mode = { "n", "v" } },
      c = { "<cmd>GoCoverage<CR>", "Tests with Coverage", mode = { "n", "v" } },
      f = { "<cmd>GoTest -f<CR>", "Test File", mode = { "n", "v" } },
      i = { "<cmd>GoGet<CR>", "Go Get", mode = { "n", "v" } },
      l = { "<cmd>GoLint<CR>", "Linter", mode = { "n", "v" } },
      p = { "<cmd>GoTestPkg<CR>", "Test Package", mode = { "n", "v" } },
      r = { "<cmd>GoRename<CR>", "Rename", mode = { "n", "v" } },
      R = { "<cmd>GoRun<CR>", "Go Run", mode = { "n", "v" } },
      t = { "<cmd>GoTest<CR>", "Test All", mode = { "n", "v" } },
      T = { "<cmd>GoDebug -t<CR>", "Debug Test", mode = { "n", "v" } },
      C = { "<cmd>GoDbgContinue<CR>", "Continue", mode = { "n", "v" } },
      d = { "<cmd>GoDebug<CR>", "File", mode = { "n", "v" } },
    },

    G = { "<cmd>LazyGit<CR>", "Lazy Git", mode = { "n", "v" } },

    ["k"] = "k9s",

    l = {
      name = "Lazy",
      c = { "<cmd>Lazy check<cr>", "Check for Updates", mode = { "n", "v" } },
      C = { "<cmd>Lazy clean<cr>", "Clean", mode = { "n", "v" } },
      d = { "<cmd>Lazy debug<cr>", "Debug", mode = { "n", "v" } },
      h = { "<cmd>Lazy home<cr>", "Home", mode = { "n", "v" } },
      H = { "<cmd>Lazy health<cr>", "Health", mode = { "n", "v" } },
      i = { "<cmd>Lazy install<cr>", "Install", mode = { "n", "v" } },
      r = { "<cmd>Lazy restore<cr>", "Restore", mode = { "n", "v" } },
      u = { "<cmd>Lazy update<cr>", "Update", mode = { "n", "v" } },
    },

    m = { "<cmd>Telescope make<cr>", "Make", mode = { "n", "v" } },

    M = {
      name = "Markdown",
      e = { "<cmd>MdEval<cr>", "Evaludate code block", mode = { "n" }},
      m = { "<cmd>MarkmapOpen<cr>", "Open Mindmap", mode = { "n" }},
      p = { "<cmd>PeekOpen<cr>", "Preview in Browser", mode = { "n" }},
      t = { "<cmd>EasyTablesCreateNew 3<cr>", "Create Table", mode = { "n" }},
    },

    n = {
      name = "Node",
      c = { "<cmd>lua require('package-info').change_version()<cr>", "Change Dependency Version", mode = { "n", "v" } },
      d = { "<cmd>lua require('package-info').delete()<cr>", "Delete Dependency", mode = { "n", "v" } },
      h = { "<cmd>lua require('package-info').hide()<cr>", "Hide Dependency Versions", mode = { "n", "v" } },
      i = { "<cmd>lua require('package-info').install()<cr>", "Install Dependency", mode = { "n", "v" } },
      s = { "<cmd>lua require('package-info').show()<cr>", "Show Dependency Versions", mode = { "n", "v" } },
      t = { "<cmd>lua require('package-info').toggle()<cr>", "Toggle Showing Dependency Versions", mode = { "n", "v" } },
      u = { "<cmd>lua require('package-info').update()<cr>", "Update Dependency", mode = { "n", "v" } },
    },

    p = {
      name = "Python",
      d = { "<cmd>lua require('dap-python').debug_selection()<cr>", "Debug Selection", mode = { "n", "v" } },
    },

    t = {
      name = "Terminal",
      f = { "<cmd>Dotenv<CR><cmd>ToggleTerm size=15 direction=float<cr>", "Floating", mode = { "n", "v" } },
      h = { "<cmd>Dotenv<CR><cmd>ToggleTerm size=15 direction=horizontal<cr>", "Horizontal", mode = { "n", "v" } },
      t = { "<cmd>Dotenv<CR><cmd>ToggleTerm direction=tab<cr>", "Tab", mode = { "n", "v" } },
      v = { "<cmd>Dotenv<CR><cmd>ToggleTerm size=70 direction=vertical<cr>", "Vertical", mode = { "n", "v" } },
    },

  }, { prefix = "," })
}

