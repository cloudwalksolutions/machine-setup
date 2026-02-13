#!/bin/zsh

# Migration script
# Migrates old dot-based naming to underscore naming

source "$(dirname "$0")/lib/common.sh"

log_info "Migrating zsh configuration files to new naming convention..."
echo ""

MIGRATED=()
SKIPPED=()

# Migrate .zshrc.aliases → .zshrc_aliases
if [[ -f "${HOME}/.zshrc.aliases" ]]; then
  if [[ -f "${HOME}/.zshrc_aliases" ]]; then
    log_warning "~/.zshrc_aliases already exists, skipping .zshrc.aliases migration"
    log_info "Please manually merge if needed: ~/.zshrc.aliases"
    SKIPPED+=(".zshrc.aliases")
  else
    cp "${HOME}/.zshrc.aliases" "${HOME}/.zshrc_aliases" || {
      log_error "Failed to migrate .zshrc.aliases"
      exit 1
    }
    log_success "Migrated .zshrc.aliases → .zshrc_aliases"
    MIGRATED+=(".zshrc.aliases")
  fi
else
  log_info ".zshrc.aliases not found (already migrated or doesn't exist)"
fi

echo ""

# Migrate .zshrc.funcs → .zshrc_funcs
if [[ -f "${HOME}/.zshrc.funcs" ]]; then
  if [[ -f "${HOME}/.zshrc_funcs" ]]; then
    log_warning "~/.zshrc_funcs already exists, skipping .zshrc.funcs migration"
    log_info "Please manually merge if needed: ~/.zshrc.funcs"
    SKIPPED+=(".zshrc.funcs")
  else
    cp "${HOME}/.zshrc.funcs" "${HOME}/.zshrc_funcs" || {
      log_error "Failed to migrate .zshrc.funcs"
      exit 1
    }
    log_success "Migrated .zshrc.funcs → .zshrc_funcs"
    MIGRATED+=(".zshrc.funcs")
  fi
else
  log_info ".zshrc.funcs not found (already migrated or doesn't exist)"
fi

echo ""

# Migrate .zshrc.secret → .zshrc_secret
if [[ -f "${HOME}/.zshrc.secret" ]]; then
  if [[ -f "${HOME}/.zshrc_secret" ]]; then
    log_warning "~/.zshrc_secret already exists, skipping .zshrc.secret migration"
    log_info "Please manually merge if needed: ~/.zshrc.secret"
    SKIPPED+=(".zshrc.secret")
  else
    cp "${HOME}/.zshrc.secret" "${HOME}/.zshrc_secret" || {
      log_error "Failed to migrate .zshrc.secret"
      exit 1
    }
    log_success "Migrated .zshrc.secret → .zshrc_secret"
    MIGRATED+=(".zshrc.secret")
  fi
else
  log_info ".zshrc.secret not found (already migrated or doesn't exist)"
fi

echo ""

# Summary
if [[ ${#MIGRATED[@]} -gt 0 ]]; then
  log_success "Migration complete! Migrated files:"
  for file in "${MIGRATED[@]}"; do
    echo "  ✅ $file"
  done
  echo ""
  log_warning "Old files have been copied to new names. The old files still exist."
  echo ""
  if confirm_action "Remove old files (.zshrc.aliases, .zshrc.funcs, .zshrc.secret)?"; then
    for file in "${MIGRATED[@]}"; do
      rm "${HOME}/${file}"
      log_info "Removed ${HOME}/${file}"
    done
    log_success "Old files removed"
  else
    log_info "Old files kept - you can manually remove them later"
  fi
else
  log_info "No files to migrate"
fi

if [[ ${#SKIPPED[@]} -gt 0 ]]; then
  echo ""
  log_warning "Skipped files (destinations already exist):"
  for file in "${SKIPPED[@]}"; do
    echo "  ⚠️  $file"
  done
fi

echo ""
log_info "Next step: Restart your terminal or run: source ~/.zshrc"
