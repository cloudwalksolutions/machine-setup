-- Custom health checks for Neovim configuration
-- Run with: :checkhealth nvim_config

local M = {}

-- Compatibility layer for different Neovim versions
local health = vim.health or require("health")

function M.check()
  health.report_start("Neovim Configuration Health")

  -- 1. Check External Dependencies
  M.check_external_deps()

  -- 2. Check Mason Tools
  M.check_mason_tools()

  -- 3. Check LSP Status
  M.check_lsp_status()

  -- 4. Check API Keys
  M.check_api_keys()

  -- 5. Check for Configuration Errors
  M.check_config_errors()

  -- 6. Check Plugin Status
  M.check_plugin_status()
end

-- Check external dependencies
function M.check_external_deps()
  health.report_start("External Dependencies")

  -- Python3 (required for many plugins)
  if vim.fn.executable("python3") == 1 then
    local python_version = vim.fn.system("python3 --version"):gsub("\n", "")
    health.report_ok("Python3: " .. python_version)
  else
    health.report_error("Python3 not found in PATH", {
      "Install Python3",
      "Ensure python3 is in your PATH"
    })
  end

  -- Node.js (for some plugins)
  if vim.fn.executable("node") == 1 then
    local node_version = vim.fn.system("node --version"):gsub("\n", "")
    health.report_ok("Node.js: " .. node_version)
  else
    health.report_warn("Node.js not found (optional for some plugins)", {
      "Install Node.js if you use plugins like Avante or Peek.nvim"
    })
  end

  -- Git (required for plugin manager)
  if vim.fn.executable("git") == 1 then
    local git_version = vim.fn.system("git --version"):gsub("\n", "")
    health.report_ok("Git: " .. git_version)
  else
    health.report_error("Git not found", {
      "Install Git - required for Lazy.nvim plugin manager"
    })
  end

  -- Deno (for Peek.nvim markdown preview)
  if vim.fn.executable("deno") == 1 then
    local deno_version = vim.fn.system("deno --version | head -1"):gsub("\n", "")
    health.report_ok("Deno: " .. deno_version)
  else
    health.report_info("Deno not found (optional for markdown preview)")
  end
end

-- Check Mason-installed tools
function M.check_mason_tools()
  health.report_start("Mason LSP Servers")

  local mason_ok, mason_registry = pcall(require, "mason-registry")
  if not mason_ok then
    health.report_warn("Mason not loaded yet", {
      "Mason will install tools on first run",
      "Open a file to trigger LSP setup"
    })
    return
  end

  local expected_servers = {
    "lua-language-server",
    "pyright",
    "gopls",
    "rust-analyzer",
    "typescript-language-server",
  }

  for _, server in ipairs(expected_servers) do
    if mason_registry.is_installed(server) then
      health.report_ok(server .. " is installed")
    else
      health.report_warn(server .. " not installed", {
        "Run :MasonInstall " .. server
      })
    end
  end
end

-- Check LSP status for current buffer
function M.check_lsp_status()
  health.report_start("LSP Status")

  local clients = vim.lsp.get_clients()
  if #clients == 0 then
    health.report_info("No LSP clients attached (open a file to trigger LSP)")
  else
    for _, client in ipairs(clients) do
      health.report_ok(string.format("LSP client: %s (id: %d)", client.name, client.id))
    end
  end
end

-- Check API keys for AI integrations
function M.check_api_keys()
  health.report_start("AI Integration API Keys")

  -- OpenAI API Key (for ChatGPT)
  local openai_key = os.getenv("OPENAI_API_KEY")
  if openai_key and #openai_key > 0 then
    health.report_ok("OPENAI_API_KEY is set")
  else
    health.report_info("OPENAI_API_KEY not set (optional for ChatGPT plugin)", {
      "Set OPENAI_API_KEY environment variable to use ChatGPT integration"
    })
  end

  -- Anthropic API Key (for Avante/Claude)
  local anthropic_key = os.getenv("ANTHROPIC_API_KEY")
  if anthropic_key and #anthropic_key > 0 then
    health.report_ok("ANTHROPIC_API_KEY is set")
  else
    health.report_info("ANTHROPIC_API_KEY not set (optional for Avante/Claude)", {
      "Set ANTHROPIC_API_KEY environment variable to use Avante integration"
    })
  end
end

-- Check for common configuration errors
function M.check_config_errors()
  health.report_start("Configuration Validation")

  -- Check Rust-tools path (common error)
  local rust_tools_ok = pcall(function()
    -- This will fail if the require path is wrong
    local plugins_content = vim.fn.readfile(vim.fn.stdpath("config") .. "/lua/core/plugins.lua")
    local has_wrong_path = false
    for _, line in ipairs(plugins_content) do
      if line:match('require%("core%.config%.lsp"%)') then
        has_wrong_path = true
        break
      end
    end
    if has_wrong_path then
      error("Found incorrect rust-tools config path")
    end
  end)

  if rust_tools_ok then
    health.report_ok("Rust-tools configuration path is correct")
  else
    health.report_error("Rust-tools uses incorrect config path", {
      "Change require(\"core.config.lsp\") to require(\"config.lsp\") in lua/core/plugins.lua"
    })
  end

  -- Check Python DAP debugpy availability
  local debugpy_paths = {
    vim.fn.expand("~/.virtualenvs/debugpy/bin/python"),
    vim.fn.expand("./venv/bin/python"),
    vim.fn.expand("./env/bin/python"),
    vim.fn.expand("./.venv/bin/python"),
  }

  local debugpy_found = false
  for _, path in ipairs(debugpy_paths) do
    if vim.fn.executable(path) == 1 then
      -- Check if debugpy is installed
      local check_cmd = path .. " -c 'import debugpy' 2>/dev/null"
      if os.execute(check_cmd) == 0 then
        health.report_ok("Python debugpy found at: " .. path)
        debugpy_found = true
        break
      end
    end
  end

  if not debugpy_found then
    health.report_warn("Python debugpy not found in common venv locations", {
      "Install debugpy: python3 -m pip install debugpy",
      "Or create venv: python3 -m venv ~/.virtualenvs/debugpy && ~/.virtualenvs/debugpy/bin/pip install debugpy"
    })
  end
end

-- Check critical plugin status
function M.check_plugin_status()
  health.report_start("Core Plugins")

  local critical_plugins = {
    { name = "lazy", module = "lazy" },
    { name = "nvim-lspconfig", module = "lspconfig" },
    { name = "nvim-cmp", module = "cmp" },
    { name = "telescope", module = "telescope" },
    { name = "nvim-treesitter", module = "nvim-treesitter" },
  }

  for _, plugin in ipairs(critical_plugins) do
    local ok, _ = pcall(require, plugin.module)
    if ok then
      health.report_ok(plugin.name .. " loaded successfully")
    else
      health.report_error(plugin.name .. " failed to load", {
        "Check :Lazy for plugin status",
        "Run :Lazy sync to update plugins"
      })
    end
  end
end

return M
