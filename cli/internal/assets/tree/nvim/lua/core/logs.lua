-- Log access utilities for debugging Neovim configuration

local M = {}

-- Open Neovim messages log
function M.open_nvim_log()
  local log_file = vim.fn.stdpath("state") .. "/log"

  if vim.fn.filereadable(log_file) == 1 then
    vim.cmd("tabnew " .. log_file)
  else
    -- Show messages in a new buffer
    vim.cmd("enew")
    vim.cmd("setlocal buftype=nofile bufhidden=wipe noswapfile")
    vim.cmd("file [Nvim Messages]")

    -- Get messages
    local messages = vim.fn.execute("messages")
    vim.api.nvim_buf_set_lines(0, 0, -1, false, vim.split(messages, "\n"))

    vim.notify("Showing recent messages (no log file found at " .. log_file .. ")", vim.log.levels.INFO)
  end
end

-- Open LSP log for current buffer's language server
function M.open_lsp_log()
  local clients = vim.lsp.get_clients({ bufnr = 0 })

  if #clients == 0 then
    vim.notify("No LSP clients attached to current buffer", vim.log.levels.WARN)
    return
  end

  -- Use first client
  local client = clients[1]
  local log_path = vim.lsp.get_log_path()

  if vim.fn.filereadable(log_path) == 1 then
    vim.cmd("tabnew " .. log_path)
    vim.notify("LSP log for " .. client.name, vim.log.levels.INFO)
  else
    vim.notify("LSP log not found at: " .. log_path, vim.log.levels.WARN)
  end
end

-- Open Mason installation log
function M.open_mason_log()
  local mason_log = vim.fn.stdpath("state") .. "/mason.log"

  if vim.fn.filereadable(mason_log) == 1 then
    vim.cmd("tabnew " .. mason_log)
  else
    vim.notify("Mason log not found at: " .. mason_log, vim.log.levels.WARN)
    vim.notify("Mason logs are usually ephemeral - check :Mason for installation status", vim.log.levels.INFO)
  end
end

-- Show plugin loading errors from Lazy.nvim
function M.show_plugin_errors()
  local lazy_ok, lazy = pcall(require, "lazy")

  if not lazy_ok then
    vim.notify("Lazy.nvim not loaded", vim.log.levels.ERROR)
    return
  end

  -- Open Lazy UI
  vim.cmd("Lazy")

  -- Show notification
  vim.notify("Check Lazy UI for plugin status and errors", vim.log.levels.INFO)
  vim.notify("Press 'l' to view logs, 'x' to view errors", vim.log.levels.INFO)
end

-- Show all log locations
function M.show_log_locations()
  local info = {
    "📋 Neovim Log Locations:",
    "",
    "Nvim State: " .. vim.fn.stdpath("state"),
    "Nvim Log: " .. vim.fn.stdpath("state") .. "/log",
    "LSP Log: " .. vim.lsp.get_log_path(),
    "Mason Log: " .. vim.fn.stdpath("state") .. "/mason.log",
    "",
    "Commands:",
    "  :NvimLog       - Open Nvim messages",
    "  :LspLog        - Open LSP log",
    "  :MasonLog      - Open Mason log",
    "  :PluginErrors  - Show plugin errors",
  }

  vim.cmd("enew")
  vim.cmd("setlocal buftype=nofile bufhidden=wipe noswapfile")
  vim.cmd("file [Log Locations]")
  vim.api.nvim_buf_set_lines(0, 0, -1, false, info)
end

return M
