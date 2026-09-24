local wk = require("which-key")

wk.setup({
  plugins = {
    marks = true,
    registers = true,
    spelling = {
      enabled = true,
      suggestions = 9,
    },
    presets = {
      operators = true,
      motions = true,
      text_objects = true,
      windows = true,
      nav = true,
      z = true,
      g = true,
    },
  },
  icons = {
    breadcrumb = "»",
    separator = "➜",
    group = "+",
  },
  layout = {
    width = { min = 20 },
    spacing = 3,
  },
  show_help = true,
})

-- SPACE PREFIX: Text Editing, LSP, File Ops, Debug (Core Native)
wk.add({
  { "<leader>t", "<cmd>lua require('nvim-toggler').toggle()<cr>", desc = "Toggle" },
  { "<leader>e", "<cmd>NvimTreeToggle<cr>", desc = "File Tree" },
  { "<leader>w", "<cmd>w<cr>", desc = "Save" },
  { "<leader>h", "<cmd>noh<cr>", desc = "Clear Highlight" },
  { "<leader>Q", "<cmd>qa<cr>", desc = "Quit All" },

  { "<leader>s", group = "Session" },
  { "<leader>sv", "<cmd>source ~/.config/nvim/init.lua<cr>", desc = "Reload Config" },

  { "<leader>l", group = "LSP" },
  { "<leader>lr", "<cmd>lua vim.lsp.buf.rename()<cr>", desc = "Rename" },
  { "<leader>la", "<cmd>lua vim.lsp.buf.code_action()<cr>", desc = "Code Action" },
  { "<leader>lk", "<cmd>lua vim.lsp.buf.signature_help()<cr>", desc = "Signature Help" },
  { "<leader>lo", "<cmd>lua vim.diagnostic.open_float()<cr>", desc = "Open Diagnostics" },
  { "<leader>ll", "<cmd>lua vim.diagnostic.setloclist()<cr>", desc = "Location List" },
  { "<leader>lD", "<cmd>lua vim.lsp.buf.type_definition()<cr>", desc = "Type Definition" },
  { "<leader>lf", "<cmd>lua vim.lsp.buf.format({ async = true })<cr>", desc = "Format", mode = { "n", "v" } },
  { "<leader>ld", "<cmd>lua vim.lsp.buf.definition()<cr>", desc = "Go to Definition" },
  { "<leader>li", "<cmd>lua vim.lsp.buf.implementation()<cr>", desc = "Go to Implementation" },
  { "<leader>lR", "<cmd>lua vim.lsp.buf.references()<cr>", desc = "References" },

  { "<leader>lw", group = "Workspace" },
  { "<leader>lwa", "<cmd>lua vim.lsp.buf.add_workspace_folder()<cr>", desc = "Add Folder" },
  { "<leader>lwr", "<cmd>lua vim.lsp.buf.remove_workspace_folder()<cr>", desc = "Remove Folder" },
  { "<leader>lwl", "<cmd>lua print(vim.inspect(vim.lsp.buf.list_workspace_folders()))<cr>", desc = "List Folders" },

  { "<leader>d", group = "Debug" },
  { "<leader>db", "<cmd>GoBreakToggle<CR>", desc = "Toggle Breakpoint", mode = { "n", "v" } },
  { "<leader>dB", "<cmd>BreakCondition<CR>", desc = "Conditional Breakpoint", mode = { "n", "v" } },
  { "<leader>dc", "<cmd>GoDbgContinue<CR>", desc = "Continue (Go)", mode = { "n", "v" } },
  { "<leader>dt", "<cmd>GoDebug -t<CR>", desc = "Debug Test (Go)", mode = { "n", "v" } },
  { "<leader>df", "<cmd>GoDebug<CR>", desc = "Debug File (Go)", mode = { "n", "v" } },
  { "<leader>dp", "<cmd>lua require('dap-python').debug_selection()<cr>", desc = "Debug Selection (Python)", mode = { "n", "v" } },

  { "gc", group = "Comments" },
})

-- COMMA PREFIX: Plugin Commands
wk.add({
  { ",C", group = "Cypress" },
  { ",Ce", "<cmd>TermExec mode=horizontal cmd='npx cypress run --browser chrome --headless --e2e'<cr>", desc = "Run E2E Tests", mode = { "n", "v" } },
  { ",Cs", "<cmd>TermExec mode=horizontal cmd='npx cypress run --spec %'<cr>", desc = "Run Spec", mode = { "n", "v" } },
  { ",Cu", "<cmd>TermExec mode=horizontal cmd='npx cypress run --browser chrome --headless --component'<cr>", desc = "Run Component Tests", mode = { "n", "v" } },

  { ",P", group = "Playwright" },
  { ",Pe", "<cmd>!npx playwright test<cr>", desc = "Run E2E Tests", mode = { "n", "v" } },
  { ",Pu", "<cmd>!npx playwright test --ui<cr>", desc = "Run UI Mode", mode = { "n", "v" } },
  { ",Pd", "<cmd>!npx playwright test --debug<cr>", desc = "Debug Tests", mode = { "n", "v" } },
  { ",Pc", "<cmd>!npx playwright codegen<cr>", desc = "Codegen", mode = { "n", "v" } },

  { ",D", "<cmd>DataViewer<CR>", desc = "Open Data Viewer", mode = { "n", "v" } },

  { ",f", group = "Search" },
  { ",fc", "<cmd>Telescope commands<CR>", desc = "Commands", mode = { "n", "v" } },
  { ",fd", "<cmd>Telescope diagnostics<CR>", desc = "Diagnostics", mode = { "n", "v" } },
  { ",ff", "<cmd>Telescope grep_string<CR>", desc = "Grep for Selected", mode = { "v" } },
  { ",fm", "<cmd>Telescope man_pages<CR>", desc = "Man Pages", mode = { "n", "v" } },
  { ",fo", "<cmd>Telescope oldfiles<CR>", desc = "Recent Files", mode = { "n", "v" } },
  { ",fr", "<cmd>Telescope registers<CR>", desc = "Registers", mode = { "n", "v" } },
  { ",fs", "<cmd>Telescope treesitter<CR>", desc = "Code", mode = { "n", "v" } },
  { ",fz", "<cmd>Telescope current_buffer_fuzzy_find<CR>", desc = "Fuzzy in Buffer", mode = { "n", "v" } },

  { ",g", group = "Go Tools" },
  { ",gg", "<cmd>GoSave<CR>", desc = "Format and Imports", mode = { "n", "v" } },
  { ",gc", "<cmd>GoCoverage<CR>", desc = "Tests with Coverage", mode = { "n", "v" } },
  { ",gf", "<cmd>GoTest -f<CR>", desc = "Test File", mode = { "n", "v" } },
  { ",gi", "<cmd>GoGet<CR>", desc = "Go Get", mode = { "n", "v" } },
  { ",gl", "<cmd>GoLint<CR>", desc = "Linter", mode = { "n", "v" } },
  { ",gp", "<cmd>GoTestPkg<CR>", desc = "Test Package", mode = { "n", "v" } },
  { ",gr", "<cmd>GoRename<CR>", desc = "Rename", mode = { "n", "v" } },
  { ",gR", "<cmd>GoRun<CR>", desc = "Go Run", mode = { "n", "v" } },
  { ",gt", "<cmd>GoTest<CR>", desc = "Test All", mode = { "n", "v" } },

  { ",G", "<cmd>LazyGit<CR>", desc = "Lazy Git", mode = { "n", "v" } },

  { ",k", desc = "k9s" },

  { ",m", "<cmd>Telescope make<cr>", desc = "Make", mode = { "n", "v" } },

  { ",M", group = "Markdown" },
  { ",Me", "<cmd>MdEval<cr>", desc = "Evaluate code block" },
  { ",Mm", "<cmd>MarkmapOpen<cr>", desc = "Open Mindmap" },
  { ",Mp", "<cmd>PeekOpen<cr>", desc = "Preview in Browser" },
  { ",Mt", "<cmd>EasyTablesCreateNew 3<cr>", desc = "Create Table" },

  { ",n", group = "Node" },
  { ",nc", "<cmd>lua require('package-info').change_version()<cr>", desc = "Change Dependency Version", mode = { "n", "v" } },
  { ",nd", "<cmd>lua require('package-info').delete()<cr>", desc = "Delete Dependency", mode = { "n", "v" } },
  { ",nh", "<cmd>lua require('package-info').hide()<cr>", desc = "Hide Dependency Versions", mode = { "n", "v" } },
  { ",ni", "<cmd>lua require('package-info').install()<cr>", desc = "Install Dependency", mode = { "n", "v" } },
  { ",ns", "<cmd>lua require('package-info').show()<cr>", desc = "Show Dependency Versions", mode = { "n", "v" } },
  { ",nt", "<cmd>lua require('package-info').toggle()<cr>", desc = "Toggle Showing Dependency Versions", mode = { "n", "v" } },
  { ",nu", "<cmd>lua require('package-info').update()<cr>", desc = "Update Dependency", mode = { "n", "v" } },

  { ",t", group = "Terminal" },
  { ",tf", "<cmd>Dotenv<CR><cmd>ToggleTerm size=15 direction=float<cr>", desc = "Floating", mode = { "n", "v" } },
  { ",th", "<cmd>Dotenv<CR><cmd>ToggleTerm size=15 direction=horizontal<cr>", desc = "Horizontal", mode = { "n", "v" } },
  { ",tt", "<cmd>Dotenv<CR><cmd>ToggleTerm direction=tab<cr>", desc = "Tab", mode = { "n", "v" } },
  { ",tv", "<cmd>Dotenv<CR><cmd>ToggleTerm size=70 direction=vertical<cr>", desc = "Vertical", mode = { "n", "v" } },

  { ",z", group = "Lazy" },
  { ",zc", "<cmd>Lazy check<cr>", desc = "Check for Updates", mode = { "n", "v" } },
  { ",zC", "<cmd>Lazy clean<cr>", desc = "Clean", mode = { "n", "v" } },
  { ",zd", "<cmd>Lazy debug<cr>", desc = "Debug", mode = { "n", "v" } },
  { ",zh", "<cmd>Lazy home<cr>", desc = "Home", mode = { "n", "v" } },
  { ",zH", "<cmd>Lazy health<cr>", desc = "Health", mode = { "n", "v" } },
  { ",zi", "<cmd>Lazy install<cr>", desc = "Install", mode = { "n", "v" } },
  { ",zr", "<cmd>Lazy restore<cr>", desc = "Restore", mode = { "n", "v" } },
  { ",zu", "<cmd>Lazy update<cr>", desc = "Update", mode = { "n", "v" } },
})
