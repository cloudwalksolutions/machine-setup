#!/bin/bash

# Zsh component script
# Handles pull/push operations for Zsh configuration

source "$(dirname "$0")/../lib/common.sh"
source "$(dirname "$0")/../lib/config.sh"

pull() {
  log_info "Pulling Zsh configuration..."

  # Validate source files exist
  validate_path "${ZSH_PATHS[zshrc_repo]}" "file" || return 1
  validate_path "${ZSH_PATHS[aliases_repo]}" "file" || return 1
  validate_path "${ZSH_PATHS[profile_repo]}" "file" || return 1
  validate_path "${ZSH_PATHS[starship_repo]}" "file" || return 1

  # Backup and copy shell rc file (deploys to ~/.zshrc or ~/.bashrc based on shell)
  local rc_local="${ZSH_PATHS[zshrc_local]}"
  if [[ "$SHELL" == *"bash"* ]] && ! command -v zsh &>/dev/null; then
    rc_local="${HOME}/.bashrc"
  fi
  if [[ -f "$rc_local" ]]; then
    backup_file "$rc_local" "zsh" || return 1
  fi
  safe_copy "${ZSH_PATHS[zshrc_repo]}" "$rc_local" "zsh" || return 1

  # Backup and copy aliases
  if [[ -f "${ZSH_PATHS[aliases_local]}" ]]; then
    backup_file "${ZSH_PATHS[aliases_local]}" "zsh" || return 1
  fi
  safe_copy "${ZSH_PATHS[aliases_repo]}" "${ZSH_PATHS[aliases_local]}" "zsh" || return 1

  # Clean up old naming convention (.zshrc.aliases → .zshrc_aliases)
  if [[ -f "${HOME}/.zshrc.aliases" ]]; then
    log_warning "Removing old .zshrc.aliases (renamed to .zshrc_aliases)"
    rm "${HOME}/.zshrc.aliases"
  fi

  # Backup and copy funcs if it exists in repo
  if [[ -f "${ZSH_PATHS[funcs_repo]}" ]]; then
    if [[ -f "${ZSH_PATHS[funcs_local]}" ]]; then
      backup_file "${ZSH_PATHS[funcs_local]}" "zsh" || return 1
    fi
    safe_copy "${ZSH_PATHS[funcs_repo]}" "${ZSH_PATHS[funcs_local]}" "zsh" || return 1
  fi

  # Backup and copy profile
  if [[ -f "${ZSH_PATHS[profile_local]}" ]]; then
    backup_file "${ZSH_PATHS[profile_local]}" "zsh" || return 1
  fi
  safe_copy "${ZSH_PATHS[profile_repo]}" "${ZSH_PATHS[profile_local]}" "zsh" || return 1

  # Create .config directory if needed for starship
  if [[ ! -d "${HOME}/.config" ]]; then
    mkdir -p "${HOME}/.config"
  fi

  # Backup and copy starship config
  if [[ -f "${ZSH_PATHS[starship_local]}" ]]; then
    backup_file "${ZSH_PATHS[starship_local]}" "zsh" || return 1
  fi
  safe_copy "${ZSH_PATHS[starship_repo]}" "${ZSH_PATHS[starship_local]}" "zsh" || return 1

  log_success "Zsh configuration pulled successfully"
  return 0
}

push() {
  log_info "Pushing Zsh configuration..."

  # Validate local files exist
  validate_path "${ZSH_PATHS[zshrc_local]}" "file" || {
    log_error "Zshrc not found at ${ZSH_PATHS[zshrc_local]}"
    return 1
  }
  validate_path "${ZSH_PATHS[aliases_local]}" "file" || {
    log_error "Zsh aliases not found at ${ZSH_PATHS[aliases_local]}"
    return 1
  }

  # Backup and copy zshrc
  if [[ -f "${ZSH_PATHS[zshrc_repo]}" ]]; then
    backup_file "${ZSH_PATHS[zshrc_repo]}" "zsh-repo" || return 1
  fi
  safe_copy "${ZSH_PATHS[zshrc_local]}" "${ZSH_PATHS[zshrc_repo]}" "zsh" || return 1

  # Backup and copy aliases
  if [[ -f "${ZSH_PATHS[aliases_repo]}" ]]; then
    backup_file "${ZSH_PATHS[aliases_repo]}" "zsh-repo" || return 1
  fi
  safe_copy "${ZSH_PATHS[aliases_local]}" "${ZSH_PATHS[aliases_repo]}" "zsh" || return 1

  # Backup and copy funcs if it exists locally
  if [[ -f "${ZSH_PATHS[funcs_local]}" ]]; then
    if [[ -f "${ZSH_PATHS[funcs_repo]}" ]]; then
      backup_file "${ZSH_PATHS[funcs_repo]}" "zsh-repo" || return 1
    fi
    safe_copy "${ZSH_PATHS[funcs_local]}" "${ZSH_PATHS[funcs_repo]}" "zsh" || return 1
  fi

  # Backup and copy profile
  if [[ -f "${ZSH_PATHS[profile_local]}" ]]; then
    if [[ -f "${ZSH_PATHS[profile_repo]}" ]]; then
      backup_file "${ZSH_PATHS[profile_repo]}" "zsh-repo" || return 1
    fi
    safe_copy "${ZSH_PATHS[profile_local]}" "${ZSH_PATHS[profile_repo]}" "zsh" || return 1
  fi

  # Backup and copy starship config
  if [[ -f "${ZSH_PATHS[starship_local]}" ]]; then
    if [[ -f "${ZSH_PATHS[starship_repo]}" ]]; then
      backup_file "${ZSH_PATHS[starship_repo]}" "zsh-repo" || return 1
    fi
    safe_copy "${ZSH_PATHS[starship_local]}" "${ZSH_PATHS[starship_repo]}" "zsh" || return 1
  fi

  log_success "Zsh configuration pushed successfully"
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
