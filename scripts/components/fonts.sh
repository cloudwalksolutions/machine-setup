#!/bin/zsh

# Fonts component script
# Handles installation of fonts (pull-only, no push needed)

source "$(dirname "$0")/../lib/common.sh"
source "$(dirname "$0")/../lib/config.sh"

pull() {
  log_info "Installing fonts..."

  # Check if we're on macOS
  if [[ "$IS_MACOS" != "true" ]]; then
    log_warning "Font installation is macOS-specific. Skipping on this platform."
    return 0
  fi

  # Validate source directory exists
  validate_path "${FONTS_PATHS[repo]}" "dir" || return 1

  # Check if fonts directory has any fonts
  if [[ -z "$(ls -A "${FONTS_PATHS[repo]}")" ]]; then
    log_warning "No fonts found in ${FONTS_PATHS[repo]}"
    return 0
  fi

  # Install fonts to system directory (requires sudo)
  log_info "Installing fonts to ${FONTS_PATHS[local]} (requires sudo)"
  sudo cp "${FONTS_PATHS[repo]}"/* "${FONTS_PATHS[local]}/" || {
    log_error "Failed to install fonts"
    return 1
  }

  log_success "Fonts installed successfully"
  return 0
}

push() {
  log_info "Fonts component does not support push operation (read-only)"
  log_info "Fonts are managed in the repository only"
  return 0
}

# Main execution
case "$1" in
  pull) pull ;;
  push) push ;;
  *)
    log_error "Usage: $0 {pull|push}"
    exit 1
    ;;
esac
