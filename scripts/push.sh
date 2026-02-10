#!/bin/zsh

# Push orchestrator script
# Pushes all component configurations from local machine to repo and commits

source "$(dirname "$0")/lib/common.sh"
source "$(dirname "$0")/lib/config.sh"

log_info "Pushing all configurations..."
echo ""

# Git pull first
log_info "Pulling latest changes from remote..."
git -C "${REPO_ROOT}" pull || {
  log_error "Failed to pull latest changes from remote"
  exit 1
}
echo ""

# Push each component (except fonts, which is read-only)
COMPONENTS=(vim zsh byobu nvim)
FAILED=()

for component in "${COMPONENTS[@]}"; do
  if ! "${REPO_ROOT}/scripts/components/${component}.sh" push; then
    FAILED+=("$component")
  fi
  echo ""  # Add spacing between components
done

if [[ ${#FAILED[@]} -gt 0 ]]; then
  log_error "Failed components: ${FAILED[*]}"
  exit 1
fi

# Show git status
echo ""
log_info "Changes to be committed:"
git -C "${REPO_ROOT}" status
echo ""

# Confirm commit and push
if ! confirm_action "Commit and push these changes to remote?"; then
  log_info "Cancelled by user"
  exit 0
fi

# Commit and push
git -C "${REPO_ROOT}" add -A
git -C "${REPO_ROOT}" commit -m "Updating remote config files" || {
  log_error "Failed to commit changes"
  exit 1
}

git -C "${REPO_ROOT}" push || {
  log_error "Failed to push to remote"
  exit 1
}

log_success "All configs synced to remote!"
