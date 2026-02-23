#!/bin/bash

# Fonts component script
# Handles installation of fonts (pull-only, no push needed)

source "$(dirname "$0")/../lib/common.sh"
source "$(dirname "$0")/../lib/config.sh"

pull() {
  log_info "Installing fonts..."

  # Validate source directory exists
  validate_path "${FONTS_PATHS[repo]}" "dir" || return 1

  # Check if fonts directory has any fonts
  if [[ -z "$(ls -A "${FONTS_PATHS[repo]}")" ]]; then
    log_warning "No fonts found in ${FONTS_PATHS[repo]}"
    return 0
  fi

  if [[ "$IS_MACOS" == "true" ]]; then
    # macOS: system font directory requires sudo
    log_info "Installing fonts to ${FONTS_PATHS[local]} (requires sudo)"
    sudo cp "${FONTS_PATHS[repo]}"/* "${FONTS_PATHS[local]}/" || {
      log_error "Failed to install fonts"
      return 1
    }
  elif [[ "$IS_LINUX" == "true" ]]; then
    # Linux: user font directory, no sudo needed
    log_info "Installing fonts to ${FONTS_PATHS[local]}"
    mkdir -p "${FONTS_PATHS[local]}"
    # Reclaim ownership if files were previously installed by root
    if [[ -n "$(find "${FONTS_PATHS[local]}" -maxdepth 1 -not -user "$(id -u)" -name '*.ttf' 2>/dev/null)" ]]; then
      sudo chown "$(id -u):$(id -g)" "${FONTS_PATHS[local]}"/*.ttf 2>/dev/null
    fi
    cp -f "${FONTS_PATHS[repo]}"/* "${FONTS_PATHS[local]}/" || {
      log_error "Failed to install fonts"
      return 1
    }
    fc-cache -f "${FONTS_PATHS[local]}" 2>/dev/null || log_warning "fc-cache failed"
  else
    log_warning "Font installation not supported on this platform. Skipping."
    return 0
  fi

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
