#!/bin/zsh

# Byobu component script
# Handles pull/push operations for Byobu/tmux configuration

source "$(dirname "$0")/../lib/common.sh"
source "$(dirname "$0")/../lib/config.sh"

pull() {
  log_info "Pulling Byobu configuration..."

  # Validate source files exist
  validate_path "${BYOBU_PATHS[bin_repo]}" "dir" || return 1
  validate_path "${BYOBU_PATHS[tmux_conf_repo]}" "file" || return 1
  validate_path "${BYOBU_PATHS[keybindings_repo]}" "file" || return 1
  validate_path "${BYOBU_PATHS[datetime_repo]}" "file" || return 1
  validate_path "${BYOBU_PATHS[statusrc_repo]}" "file" || return 1

  # Create .byobu directories if they don't exist
  mkdir -p "${HOME}/.byobu/bin" || {
    log_error "Failed to create ~/.byobu/bin directory"
    return 1
  }

  # Backup and copy bin directory
  if [[ -d "${BYOBU_PATHS[bin_local]}" ]] && [[ -n "$(ls -A "${BYOBU_PATHS[bin_local]}")" ]]; then
    backup_file "${BYOBU_PATHS[bin_local]}" "byobu" || return 1
  fi
  cp -r "${BYOBU_PATHS[bin_repo]}"/* "${BYOBU_PATHS[bin_local]}/" || {
    log_error "Failed to copy byobu bin files"
    return 1
  }

  # Backup and copy tmux.conf
  if [[ -f "${BYOBU_PATHS[tmux_conf_local]}" ]]; then
    backup_file "${BYOBU_PATHS[tmux_conf_local]}" "byobu" || return 1
  fi
  safe_copy "${BYOBU_PATHS[tmux_conf_repo]}" "${BYOBU_PATHS[tmux_conf_local]}" "byobu" || return 1

  # Backup and copy keybindings
  if [[ -f "${BYOBU_PATHS[keybindings_local]}" ]]; then
    backup_file "${BYOBU_PATHS[keybindings_local]}" "byobu" || return 1
  fi
  safe_copy "${BYOBU_PATHS[keybindings_repo]}" "${BYOBU_PATHS[keybindings_local]}" "byobu" || return 1

  # Backup and copy datetime
  if [[ -f "${BYOBU_PATHS[datetime_local]}" ]]; then
    backup_file "${BYOBU_PATHS[datetime_local]}" "byobu" || return 1
  fi
  safe_copy "${BYOBU_PATHS[datetime_repo]}" "${BYOBU_PATHS[datetime_local]}" "byobu" || return 1

  # Backup and copy statusrc
  if [[ -f "${BYOBU_PATHS[statusrc_local]}" ]]; then
    backup_file "${BYOBU_PATHS[statusrc_local]}" "byobu" || return 1
  fi
  safe_copy "${BYOBU_PATHS[statusrc_repo]}" "${BYOBU_PATHS[statusrc_local]}" "byobu" || return 1

  log_success "Byobu configuration pulled successfully"
  return 0
}

push() {
  log_info "Pushing Byobu configuration..."

  # Validate local directory exists
  validate_path "${HOME}/.byobu" "dir" || {
    log_error "Byobu config not found at ${HOME}/.byobu"
    return 1
  }

  # Backup and copy bin directory
  if [[ -d "${BYOBU_PATHS[bin_repo]}" ]]; then
    backup_file "${BYOBU_PATHS[bin_repo]}" "byobu-repo" || return 1
  fi
  mkdir -p "${BYOBU_PATHS[bin_repo]}"
  if [[ -d "${BYOBU_PATHS[bin_local]}" ]] && [[ -n "$(ls -A "${BYOBU_PATHS[bin_local]}")" ]]; then
    cp -r "${BYOBU_PATHS[bin_local]}"/* "${BYOBU_PATHS[bin_repo]}/" || {
      log_error "Failed to copy byobu bin files"
      return 1
    }
  fi

  # Backup and copy tmux.conf
  if [[ -f "${BYOBU_PATHS[tmux_conf_local]}" ]]; then
    if [[ -f "${BYOBU_PATHS[tmux_conf_repo]}" ]]; then
      backup_file "${BYOBU_PATHS[tmux_conf_repo]}" "byobu-repo" || return 1
    fi
    safe_copy "${BYOBU_PATHS[tmux_conf_local]}" "${BYOBU_PATHS[tmux_conf_repo]}" "byobu" || return 1
  fi

  # Backup and copy keybindings
  if [[ -f "${BYOBU_PATHS[keybindings_local]}" ]]; then
    if [[ -f "${BYOBU_PATHS[keybindings_repo]}" ]]; then
      backup_file "${BYOBU_PATHS[keybindings_repo]}" "byobu-repo" || return 1
    fi
    safe_copy "${BYOBU_PATHS[keybindings_local]}" "${BYOBU_PATHS[keybindings_repo]}" "byobu" || return 1
  fi

  # Backup and copy datetime
  if [[ -f "${BYOBU_PATHS[datetime_local]}" ]]; then
    if [[ -f "${BYOBU_PATHS[datetime_repo]}" ]]; then
      backup_file "${BYOBU_PATHS[datetime_repo]}" "byobu-repo" || return 1
    fi
    safe_copy "${BYOBU_PATHS[datetime_local]}" "${BYOBU_PATHS[datetime_repo]}" "byobu" || return 1
  fi

  # Backup and copy statusrc
  if [[ -f "${BYOBU_PATHS[statusrc_local]}" ]]; then
    if [[ -f "${BYOBU_PATHS[statusrc_repo]}" ]]; then
      backup_file "${BYOBU_PATHS[statusrc_repo]}" "byobu-repo" || return 1
    fi
    safe_copy "${BYOBU_PATHS[statusrc_local]}" "${BYOBU_PATHS[statusrc_repo]}" "byobu" || return 1
  fi

  log_success "Byobu configuration pushed successfully"
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
