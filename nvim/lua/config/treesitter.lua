local ts = require("nvim-treesitter")

if vim.fn.executable("tree-sitter") == 1 then
  ts.install({ "go", "python", "cpp", "lua", "vim", "json", "toml", "yaml", "html", "css", "javascript", "typescript", "rust", "bash", "dockerfile" })
end

vim.api.nvim_create_autocmd("FileType", {
  group = vim.api.nvim_create_augroup("treesitter_start", { clear = true }),
  callback = function(args)
    if pcall(vim.treesitter.start, args.buf) then
      vim.bo[args.buf].indentexpr = "v:lua.require'nvim-treesitter'.indentexpr()"
    end
  end,
})
