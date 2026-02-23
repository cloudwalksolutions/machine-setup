#!/bin/bash

# Init orchestrator script
# Sets up a fresh machine with all configurations and dependencies

source "$(dirname "$0")/lib/common.sh"
source "$(dirname "$0")/lib/config.sh"

log_info "Initializing machine setup..."
echo ""

# Install dependencies
if [[ "$IS_MACOS" == "true" ]]; then
  log_info "Installing Homebrew packages..."
  brew install byobu || log_warning "Failed to install byobu"
  brew install nvim || log_warning "Failed to install nvim"
elif [[ "$IS_LINUX" == "true" ]]; then
  log_info "Installing packages via apt..."
  sudo apt update || log_warning "Failed to update apt"
  sudo apt install -y byobu zsh || log_warning "Failed to install packages"
  # Neovim: apt version is too old (need 0.10+), use AppImage
  nvim_minor=$(nvim --version 2>/dev/null | head -1 | sed 's/.*v[0-9]*\.\([0-9]*\).*/\1/')
  if ! command -v nvim &>/dev/null || [[ "${nvim_minor:-0}" -lt 10 ]]; then
    log_info "Installing Neovim AppImage (0.10+)..."
    nvim_arch="x86_64"
    [[ "$(uname -m)" == "aarch64" ]] && nvim_arch="aarch64"
    nvim_dest="${HOME}/.local/bin/nvim"
    mkdir -p "$(dirname "$nvim_dest")"
    curl -fsSL "https://github.com/neovim/neovim/releases/download/v0.11.6/nvim-linux-${nvim_arch}.appimage" -o "$nvim_dest" \
      && chmod +x "$nvim_dest" \
      || log_warning "Failed to install Neovim AppImage"
  fi
  # starship (prompt fallback for bash, or zsh without oh-my-zsh)
  if ! command -v starship &>/dev/null; then
    log_info "Installing starship prompt..."
    curl -sS https://starship.rs/install.sh | sh -s -- -y || log_warning "Failed to install starship"
  fi
else
  log_warning "Unsupported platform. Skipping package installation."
fi
echo ""

# Set zsh as default shell on Linux
if [[ "$IS_LINUX" == "true" ]] && command -v zsh &>/dev/null; then
  if [[ "$SHELL" != *"zsh"* ]]; then
    log_info "Setting zsh as default shell..."
    chsh -s "$(which zsh)" || log_warning "Failed to set zsh as default shell (you can run: chsh -s \$(which zsh))"
  else
    log_info "zsh is already the default shell"
  fi
  echo ""
fi

# Install fonts
log_info "Installing fonts..."
if [[ "$IS_MACOS" == "true" ]]; then
  sudo cp "${FONTS_PATHS[repo]}"/* "${FONTS_PATHS[local]}" 2>/dev/null || log_warning "Failed to install fonts"
elif [[ "$IS_LINUX" == "true" ]]; then
  mkdir -p "${FONTS_PATHS[local]}"
  cp -f "${FONTS_PATHS[repo]}"/* "${FONTS_PATHS[local]}/" 2>/dev/null || log_warning "Failed to install fonts"
  fc-cache -f "${FONTS_PATHS[local]}" 2>/dev/null || log_warning "fc-cache not found or failed"
fi
echo ""

# Optional: Install completion files (macOS only)
if [[ "$IS_MACOS" == "true" ]]; then
  if confirm_action "Install zsh/bash completion files to /etc/ (requires sudo)?"; then
    sudo cp "${REPO_ROOT}/zsh/zsh_completion.d" /etc/zsh_completion.d 2>/dev/null
    sudo cp "${REPO_ROOT}/zsh/bash_completion.d" /etc/bash_completion.d 2>/dev/null
    echo ""
  fi
fi

# Initialize secret template if it doesn't exist
if [[ ! -f "${HOME}/.zshrc_secret" ]]; then
  log_info "Initializing .zshrc_secret template..."
  if [[ -f "${REPO_ROOT}/zsh/zshrc_secret.template" ]]; then
    cp "${REPO_ROOT}/zsh/zshrc_secret.template" "${HOME}/.zshrc_secret"
    log_success "Created ~/.zshrc_secret from template"
    log_warning "Don't forget to add your personal secrets to ~/.zshrc_secret"
  else
    log_warning "Secret template not found, skipping"
  fi
  echo ""
fi

# Pull all configs
log_info "Pulling configurations..."
echo ""
"${REPO_ROOT}/scripts/pull.sh"

echo ""
log_success "Machine setup complete!"
echo ""
log_info "Final steps:"
echo "  1. Edit ~/.zshrc_secret and add your API keys/secrets"
echo "  2. Restart your terminal"
echo "  3. Open nvim to let plugins install"
