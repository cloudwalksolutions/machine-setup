#!/bin/bash

# Vim component script
# Handles pull/push operations for Vim configuration

source "$(dirname "$0")/../lib/common.sh"
source "$(dirname "$0")/../lib/config.sh"

pull() {
  log_info "Pulling Vim configuration..."

  # Validate source files exist
  validate_path "${VIM_PATHS[vimrc_repo]}" "file" || return 1
  validate_path "${VIM_PATHS[colors_repo]}" "file" || return 1

  # Create colors directory if it doesn't exist (fixes bug from original copy.sh line 6)
  if [[ ! -d "${HOME}/.vim/colors" ]]; then
    mkdir -p "${HOME}/.vim/colors" || {
      log_error "Failed to create ~/.vim/colors directory"
      return 1
    }
  fi

  # Backup and copy vimrc
  if [[ -f "${VIM_PATHS[vimrc_local]}" ]]; then
    backup_file "${VIM_PATHS[vimrc_local]}" "vim" || return 1
  fi
  safe_copy "${VIM_PATHS[vimrc_repo]}" "${VIM_PATHS[vimrc_local]}" "vim" || return 1

  # Backup and copy color scheme
  if [[ -f "${VIM_PATHS[colors_local]}" ]]; then
    backup_file "${VIM_PATHS[colors_local]}" "vim" || return 1
  fi
  safe_copy "${VIM_PATHS[colors_repo]}" "${VIM_PATHS[colors_local]}" "vim" || return 1

  log_success "Vim configuration pulled successfully"
  return 0
}

push() {
  log_info "Pushing Vim configuration..."

  # Validate local files exist
  validate_path "${VIM_PATHS[vimrc_local]}" "file" || {
    log_error "Vimrc not found at ${VIM_PATHS[vimrc_local]}"
    return 1
  }
  validate_path "${VIM_PATHS[colors_local]}" "file" || {
    log_error "Vim color scheme not found at ${VIM_PATHS[colors_local]}"
    return 1
  }

  # Backup and copy vimrc
  if [[ -f "${VIM_PATHS[vimrc_repo]}" ]]; then
    backup_file "${VIM_PATHS[vimrc_repo]}" "vim-repo" || return 1
  fi
  safe_copy "${VIM_PATHS[vimrc_local]}" "${VIM_PATHS[vimrc_repo]}" "vim" || return 1

  # Backup and copy color scheme
  if [[ -f "${VIM_PATHS[colors_repo]}" ]]; then
    backup_file "${VIM_PATHS[colors_repo]}" "vim-repo" || return 1
  fi
  safe_copy "${VIM_PATHS[colors_local]}" "${VIM_PATHS[colors_repo]}" "vim" || return 1

  log_success "Vim configuration pushed successfully"
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
