#!/bin/bash

# Init orchestrator script
# Sets up a fresh machine with all configurations and dependencies

source "$(dirname "$0")/lib/common.sh"
source "$(dirname "$0")/lib/config.sh"

log_info "Initializing machine setup..."
echo ""

# Install Homebrew dependencies
log_info "Installing Homebrew packages..."
brew install starship || log_warning "Failed to install starship"
brew install byobu || log_warning "Failed to install byobu"
brew install nvim || log_warning "Failed to install nvim"
echo ""

# Install fonts (system-wide)
log_info "Installing fonts..."
sudo cp "${REPO_ROOT}/fonts"/* /Library/Fonts 2>/dev/null || log_warning "Failed to install fonts"
echo ""

# Optional: Install completion files
if confirm_action "Install zsh/bash completion files to /etc/ (requires sudo)?"; then
  sudo cp "${REPO_ROOT}/zsh/zsh_completion.d" /etc/zsh_completion.d 2>/dev/null
  sudo cp "${REPO_ROOT}/zsh/bash_completion.d" /etc/bash_completion.d 2>/dev/null
  echo ""
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
