-- Python DAP configuration with auto-detection of debugpy

-- Auto-detect Python with debugpy from common venv locations
local function find_python_with_debugpy()
  local python_paths = {
    vim.fn.stdpath("data") .. "/mason/packages/debugpy/venv/bin/python",
    vim.fn.expand("~/.virtualenvs/debugpy/bin/python"),
    vim.fn.expand("./venv/bin/python"),
    vim.fn.expand("./env/bin/python"),
    vim.fn.expand("./.venv/bin/python"),
    vim.fn.exepath("python3"),  -- System python3 as fallback
  }

  for _, path in ipairs(python_paths) do
    if vim.fn.executable(path) == 1 then
      -- Check if debugpy is available
      local check_cmd = path .. " -c 'import debugpy' 2>/dev/null"
      if os.execute(check_cmd) == 0 then
        return path
      end
    end
  end

  return nil
end

local python_path = find_python_with_debugpy()

if python_path then
  require('dap-python').setup(python_path)
else
  vim.notify("debugpy not found in any venv - Python debugging won't work", vim.log.levels.DEBUG)
  vim.notify("Install with: python3 -m pip install debugpy", vim.log.levels.DEBUG)
end
