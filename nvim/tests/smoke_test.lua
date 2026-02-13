-- Minimal smoke tests for Neovim configuration
-- Run headless: nvim --headless -l nvim/tests/smoke_test.lua

-- Set up Lua path to find modules in nvim config
local config_path = vim.fn.stdpath("config")
package.path = config_path .. "/lua/?.lua;" .. config_path .. "/lua/?/init.lua;" .. package.path

local failed_tests = {}
local passed_tests = {}

local function test(name, fn)
  local ok, err = pcall(fn)
  if ok then
    table.insert(passed_tests, name)
    print("✅ PASS: " .. name)
  else
    table.insert(failed_tests, name)
    print("❌ FAIL: " .. name)
    print("   Error: " .. tostring(err))
  end
end

-- Test 1: Config loads without Lua errors
test("Config loads without Lua errors", function()
  vim.cmd("source " .. vim.fn.stdpath("config") .. "/init.lua")
  assert(true, "Config loaded")
end)

-- Test 2: Critical plugins are available
test("Critical plugins exist", function()
  local critical = {
    { name = "lazy.nvim", module = "lazy" },
    { name = "nvim-lspconfig", module = "lspconfig" },
    { name = "nvim-cmp", module = "cmp" },
    { name = "telescope.nvim", module = "telescope" },
    { name = "nvim-treesitter", module = "nvim-treesitter" },
    { name = "mason.nvim", module = "mason" },
    { name = "mason-lspconfig", module = "mason-lspconfig" },
    { name = "gitsigns", module = "gitsigns" },
    { name = "nvim-tree", module = "nvim-tree" },
    { name = "which-key", module = "which-key" },
  }

  for _, plugin in ipairs(critical) do
    local ok = pcall(require, plugin.module)
    assert(ok, plugin.name .. " not found")
  end
end)

-- Test 3: LSP configuration exists
test("LSP configuration is valid", function()
  local lsp_config = require("config.lsp")
  assert(lsp_config.on_attach ~= nil, "LSP on_attach not defined")
  assert(lsp_config.capabilities ~= nil, "LSP capabilities not defined")
end)

-- Test 4: Health check module exists
test("Health check module exists", function()
  local health = require("health.nvim_config")
  assert(health.check ~= nil, "Health check function not defined")
end)

-- Test 5: Log utilities exist
test("Log utilities exist", function()
  local logs = require("core.logs")
  assert(logs.open_nvim_log ~= nil, "open_nvim_log not defined")
  assert(logs.open_lsp_log ~= nil, "open_lsp_log not defined")
end)

-- Test 6: No plugin configuration errors
test("Plugin configurations are valid", function()
  -- Capture any errors during startup
  local errors = {}
  local old_notify = vim.notify
  vim.notify = function(msg, level)
    if level == vim.log.levels.ERROR then
      table.insert(errors, msg)
    end
  end

  -- Wait a moment for plugins to load
  vim.wait(1000)

  vim.notify = old_notify

  assert(#errors == 0, "Found " .. #errors .. " configuration errors: " .. table.concat(errors, "; "))
end)

-- Summary
print("\n" .. string.rep("=", 50))
print("Test Summary:")
print("  Passed: " .. #passed_tests)
print("  Failed: " .. #failed_tests)
print(string.rep("=", 50))

if #failed_tests > 0 then
  print("\nFailed tests:")
  for _, name in ipairs(failed_tests) do
    print("  - " .. name)
  end
  os.exit(1)
else
  print("\n🎉 All tests passed!")
  os.exit(0)
end
