local package_info = require("package-info")
local notify = require("notify")

local get_status = function()
  return require('package-info').get_status()
end

local get_package_info = function()
  notify(get_status(), "warning", {title = "Fetching latest versions"})
  while get_status() ~= "" do
    print("Getting status")
  end

  notify(result, "info", {title = "Latest versions found"})
end

-- vim.api.nvim_create_autocmd("BufEnter", {
--     pattern = "package.json",
--     callback = function()
--         get_package_info()
--     end
-- })

package_info.setup({
    highlights = {
        up_to_date = { fg = "#3C4048" },
        outdated = { fg = "#d19a66" },
    },
    icons = {
        enable = true,
        style = {
            up_to_date = "|  ",
            outdated = "|  ",
        },
    },
    autostart = true,
    hide_up_to_date = true,
    hide_unstable_versions = false,
    package_manager = 'yarn'
})

