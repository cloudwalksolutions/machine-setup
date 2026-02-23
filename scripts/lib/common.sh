#!/bin/bash

# Common utilities for machine-setup scripts
# Provides backup, logging, validation, and confirmation functions

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get repo root directory
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKUP_DIR="${REPO_ROOT}/backups"
BACKUP_RETENTION=5

# Logging functions
log_info() {
  echo "${BLUE}ℹ️  $1${NC}"
}

log_error() {
  echo "${RED}❌ ERROR: $1${NC}" >&2
}

log_success() {
  echo "${GREEN}✅ $1${NC}"
}

log_warning() {
  echo "${YELLOW}⚠️  WARNING: $1${NC}"
}

# Validate path exists and is of correct type
# Usage: validate_path <path> <type>
# type can be: file, dir, or any (default)
validate_path() {
  local path="$1"
  local type="${2:-any}"

  if [[ ! -e "$path" ]]; then
    log_error "Path does not exist: $path"
    return 1
  fi

  case "$type" in
    file)
      if [[ ! -f "$path" ]]; then
        log_error "Path is not a file: $path"
        return 1
      fi
      ;;
    dir)
      if [[ ! -d "$path" ]]; then
        log_error "Path is not a directory: $path"
        return 1
      fi
      ;;
  esac

  return 0
}

# Create versioned backup of a file or directory
# Usage: backup_file <source_path> <component_name>
backup_file() {
  local source="$1"
  local component="$2"
  local component_backup_dir="${BACKUP_DIR}/${component}"

  # Check if source exists
  if [[ ! -e "$source" ]]; then
    log_warning "Backup skipped: source does not exist: $source"
    return 0
  fi

  # Create component backup directory if it doesn't exist
  mkdir -p "$component_backup_dir" || {
    log_error "Failed to create backup directory: $component_backup_dir"
    return 1
  }

  # Find next version number
  local version=1
  if [[ -d "$component_backup_dir" ]]; then
    # Get list of existing version directories
    local existing_versions=($(ls -1 "$component_backup_dir" 2>/dev/null | grep -E '^v[0-9]+$' | sed 's/^v//' | sort -n))
    if [[ ${#existing_versions[@]} -gt 0 ]]; then
      local last_version=${existing_versions[-1]}
      version=$((last_version + 1))
    fi
  fi

  local backup_path="${component_backup_dir}/v${version}"

  # Create versioned backup directory
  mkdir -p "$backup_path" || {
    log_error "Failed to create backup directory: $backup_path"
    return 1
  }

  # Copy to backup
  if [[ -d "$source" ]]; then
    cp -r "$source" "$backup_path/" || {
      log_error "Failed to backup directory: $source"
      return 1
    }
  else
    cp "$source" "$backup_path/" || {
      log_error "Failed to backup file: $source"
      return 1
    }
  fi

  log_info "Backed up to: $backup_path"
  return 0
}

# Safe copy with validation and backup
# Usage: safe_copy <src> <dest> <component_name>
safe_copy() {
  local src="$1"
  local dest="$2"
  local component="$3"

  # Validate source exists
  validate_path "$src" || return 1

  # Backup destination if it exists
  if [[ -e "$dest" ]]; then
    backup_file "$dest" "$component" || return 1
  fi

  # Create parent directory if needed
  local parent_dir="$(dirname "$dest")"
  if [[ ! -d "$parent_dir" ]]; then
    mkdir -p "$parent_dir" || {
      log_error "Failed to create parent directory: $parent_dir"
      return 1
    }
  fi

  # Copy
  if [[ -d "$src" ]]; then
    cp -r "$src" "$dest" || {
      log_error "Failed to copy directory $src to $dest"
      return 1
    }
  else
    cp "$src" "$dest" || {
      log_error "Failed to copy file $src to $dest"
      return 1
    }
  fi

  return 0
}

# List all backups for a component
# Usage: list_backups <component_name>
list_backups() {
  local component="$1"
  local component_backup_dir="${BACKUP_DIR}/${component}"

  if [[ ! -d "$component_backup_dir" ]]; then
    log_info "No backups found for $component"
    return 0
  fi

  local backups=($(ls -1 "$component_backup_dir" 2>/dev/null | grep -E '^v[0-9]+$' | sort -V))

  if [[ ${#backups[@]} -eq 0 ]]; then
    log_info "No backups found for $component"
    return 0
  fi

  echo "${BLUE}Backups for $component:${NC}"
  for backup in "${backups[@]}"; do
    echo "  - $backup"
  done

  return 0
}

# List all backups across all components
list_all_backups() {
  log_info "Listing all backups..."
  echo ""

  if [[ ! -d "$BACKUP_DIR" ]]; then
    log_info "No backups directory found"
    return 0
  fi

  # Get list of component directories
  local components=($(ls -1 "$BACKUP_DIR" 2>/dev/null))

  for component in "${components[@]}"; do
    # Skip .gitkeep
    [[ "$component" == ".gitkeep" ]] && continue

    list_backups "$component"
    echo ""
  done
}

# Prompt user for confirmation
# Usage: confirm_action "Prompt message"
# Returns: 0 if yes, 1 if no
confirm_action() {
  local prompt="$1"
  echo -n "${YELLOW}$prompt (y/n): ${NC}"
  read -r response

  case "$response" in
    [yY]|[yY][eE][sS])
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

# Backup all local configs manually
backup_all() {
  log_info "Creating manual backup of all local configurations..."

  # Backup each component if it exists
  [[ -d "${HOME}/.config/nvim" ]] && backup_file "${HOME}/.config/nvim" "nvim-manual"
  [[ -f "${HOME}/.zshrc" ]] && backup_file "${HOME}/.zshrc" "zsh-manual"
  [[ -f "${HOME}/.zshrc_aliases" ]] && backup_file "${HOME}/.zshrc_aliases" "zsh-manual"
  [[ -f "${HOME}/.profile" ]] && backup_file "${HOME}/.profile" "zsh-manual"
  [[ -d "${HOME}/.byobu" ]] && backup_file "${HOME}/.byobu" "byobu-manual"
  [[ -f "${HOME}/.vimrc" ]] && backup_file "${HOME}/.vimrc" "vim-manual"

  log_success "Manual backup complete"
  echo ""
  log_info "To view backups, run: make backups-list"
}

# Restore a specific backup version
# Usage: restore_backup <component_name> <version>
restore_backup() {
  local component="$1"
  local version="$2"
  local backup_path="${BACKUP_DIR}/${component}/v${version}"

  if [[ ! -d "$backup_path" ]]; then
    log_error "Backup version v${version} not found for $component"
    return 1
  fi

  log_warning "This will overwrite your current $component configuration"
  if ! confirm_action "Restore $component from v${version}?"; then
    log_info "Cancelled"
    return 0
  fi

  # Component-specific restore logic would go here
  log_info "Restoring $component from v${version}..."
  log_info "Manual restore: cp -r $backup_path/* <destination>"

  return 0
}
