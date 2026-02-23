#!/bin/bash

# Neovim component script
# Handles pull/push operations for Neovim configuration

source "$(dirname "$0")/../lib/common.sh"
source "$(dirname "$0")/../lib/config.sh"

pull() {
  log_info "Pulling Neovim configuration..."

  # Validate source exists
  validate_path "${NVIM_PATHS[repo]}" "dir" || return 1
  validate_path "${NVIM_PATHS[monokai_repo]}" "file" || return 1

  # Backup existing nvim config if present
  if [[ -d "${NVIM_PATHS[local]}" ]]; then
    backup_file "${NVIM_PATHS[local]}" "nvim" || return 1
  fi

  # Remove old config
  rm -rf "${NVIM_PATHS[local]}"

  # Create .config directory if needed
  mkdir -p "${HOME}/.config" || {
    log_error "Failed to create ~/.config directory"
    return 1
  }

  # Copy nvim directory from repo to local
  cp -r "${NVIM_PATHS[repo]}" "${NVIM_PATHS[local]}" || {
    log_error "Failed to copy Neovim config"
    return 1
  }

  # Handle monokai theme
  # Create monokai directory if it doesn't exist
  if [[ ! -d "${NVIM_PATHS[monokai_local]}" ]]; then
    mkdir -p "${NVIM_PATHS[monokai_local]}" || {
      log_error "Failed to create monokai directory"
      return 1
    }
  fi

  # Backup and copy monokai.lua
  if [[ -f "${NVIM_PATHS[monokai_local]}/monokai.lua" ]]; then
    backup_file "${NVIM_PATHS[monokai_local]}/monokai.lua" "nvim" || return 1
  fi
  safe_copy "${NVIM_PATHS[monokai_repo]}" "${NVIM_PATHS[monokai_local]}/monokai.lua" "nvim" || return 1

  log_success "Neovim configuration pulled successfully"
  return 0
}

push() {
  log_info "Pushing Neovim configuration..."

  # Validate local config exists
  validate_path "${NVIM_PATHS[local]}" "dir" || {
    log_error "Neovim config not found at ${NVIM_PATHS[local]}"
    return 1
  }

  # Backup repo version if exists
  if [[ -d "${NVIM_PATHS[repo]}" ]]; then
    backup_file "${NVIM_PATHS[repo]}" "nvim-repo" || return 1
  fi

  # Remove old repo version
  rm -rf "${NVIM_PATHS[repo]}"

  # Copy from local to repo
  cp -r "${NVIM_PATHS[local]}" "${NVIM_PATHS[repo]}" || {
    log_error "Failed to copy Neovim config"
    return 1
  }

  # Handle monokai theme
  if [[ -f "${NVIM_PATHS[monokai_local]}/monokai.lua" ]]; then
    if [[ -f "${NVIM_PATHS[monokai_repo]}" ]]; then
      backup_file "${NVIM_PATHS[monokai_repo]}" "nvim-repo" || return 1
    fi
    safe_copy "${NVIM_PATHS[monokai_local]}/monokai.lua" "${NVIM_PATHS[monokai_repo]}" "nvim" || return 1
  fi

  log_success "Neovim configuration pushed successfully"
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
