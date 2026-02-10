#!/bin/zsh

# Pull orchestrator script
# Pulls all component configurations from repo to local machine

source "$(dirname "$0")/lib/common.sh"
source "$(dirname "$0")/lib/config.sh"

COMPONENTS=(vim zsh byobu nvim fonts)
FAILED=()

log_info "Pulling all configurations..."
echo ""

for component in "${COMPONENTS[@]}"; do
  if ! "${REPO_ROOT}/scripts/components/${component}.sh" pull; then
    FAILED+=("$component")
  fi
  echo ""  # Add spacing between components
done

if [[ ${#FAILED[@]} -eq 0 ]]; then
  log_success "All components pulled successfully!"
  echo ""
  log_info "Next steps:"
  echo "  - Restart your terminal or run: source ~/.zshrc"
  echo "  - Open nvim to let plugins install"
  exit 0
else
  log_error "Failed components: ${FAILED[*]}"
  exit 1
fi
