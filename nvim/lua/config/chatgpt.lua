-- ChatGPT configuration
-- Requires OPENAI_API_KEY environment variable
-- Set in ~/.zshrc_secret: export OPENAI_API_KEY="your-key-here"

local api_key = os.getenv("OPENAI_API_KEY")

if api_key then
  require("chatgpt").setup({
    api_key_cmd = "echo " .. vim.fn.shellescape(api_key)
  })
else
  vim.notify("OPENAI_API_KEY not set - ChatGPT plugin won't work", vim.log.levels.WARN)
end
