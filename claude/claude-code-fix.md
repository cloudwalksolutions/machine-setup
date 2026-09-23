# claude-code-fix.md

## Fix Neovim Treesitter Error: Missing `range` Method

### Problem:
The error `attempt to call method 'range' (a nil value)` in `treesitter.lua` indicates **incompatible Neovim version** or **outdated Treesitter plugin**.

---

### Step 1: Verify Neovim Version
```bash
nvim --version
```
- **Expected**: `NVIM v0.13+`
- **If outdated**: Proceed to Step 2.

---

### Step 2: Reinstall Neovim via Homebrew
```bash
# Remove old Neovim
brew uninstall neovim
rm -rf /opt/homebrew/Cellar/neovim
rm -rf /usr/local/bin/nvim

# Install latest version
brew tap neovim/neovim
brew install neovim
```

---

### Step 3: Verify Installation
```bash
nvim --version
```
- **Expected**: `NVIM v0.13+`
- If it still shows `0.12.5`, skip to **Step 4**.

---

### Step 4: Manual Neovim Install (If Homebrew Fails)
1. **Download Neovim** from [https://neovim.io/download](https://neovim.io/download).
2. **Replace the binary**:
   ```bash
   sudo mv /usr/local/bin/nvim /usr/local/bin/nvim-old
   sudo ln -s /path/to/downloaded/nvim /usr/local/bin/nvim
   ```
3. **Verify**:
   ```bash
   nvim --version
   ```

---

### Step 5: Reinstall Treesitter and Plugins
```bash
nvim -u nvim/init.lua --headless -c ":TSInstall all" -c "q"
```

---

### Step 6: Validate Fix
```bash
make test-nvim
```
- **Expected**: All tests pass with no errors.

---

### Notes
- If Homebrew fails, ensure no conflicts exist with `/usr/bin/nvim`.
- Always use `brew --prefix neovim` to confirm the correct path.

---

### Done
This ensures Neovim is updated, Treesitter is compatible, and your environment is validated.