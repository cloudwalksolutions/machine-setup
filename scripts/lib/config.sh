#!/bin/bash

# Configuration constants for machine-setup scripts
# Defines all file paths and component mappings

# Repo root directory (auto-detected from scripts/lib/, works in both bash and zsh)
if [[ -n "${BASH_SOURCE[0]:-}" ]]; then
  REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
else
  REPO_ROOT="$(cd "$(dirname "${(%):-%x}")/../.." && pwd)"
fi

# Backup configuration
export BACKUP_DIR="${REPO_ROOT}/backups"
export BACKUP_RETENTION=5

# Neovim paths
declare -gA NVIM_PATHS=(
  [repo]="${REPO_ROOT}/nvim"
  [local]="${HOME}/.config/nvim"
  [monokai_repo]="${REPO_ROOT}/monokai.lua"
  [monokai_local]="${HOME}/.local/share/nvim/site/pack/packer/start/monokai.nvim/lua"
)

# Zsh paths
declare -gA ZSH_PATHS=(
  [zshrc_repo]="${REPO_ROOT}/zsh/zshrc"
  [zshrc_local]="${HOME}/.zshrc"
  [aliases_repo]="${REPO_ROOT}/zsh/zshrc_aliases"
  [aliases_local]="${HOME}/.zshrc_aliases"
  [funcs_repo]="${REPO_ROOT}/zsh/zshrc_funcs"
  [funcs_local]="${HOME}/.zshrc_funcs"
  [profile_repo]="${REPO_ROOT}/zsh/profile"
  [profile_local]="${HOME}/.profile"
  [starship_repo]="${REPO_ROOT}/zsh/starship.toml"
  [starship_local]="${HOME}/.config/starship.toml"
)

# Byobu paths
declare -gA BYOBU_PATHS=(
  [bin_repo]="${REPO_ROOT}/byobu/bin"
  [bin_local]="${HOME}/.byobu/bin"
  [tmux_conf_repo]="${REPO_ROOT}/byobu/.tmux.conf"
  [tmux_conf_local]="${HOME}/.byobu/.tmux.conf"
  [keybindings_repo]="${REPO_ROOT}/byobu/keybindings.tmux"
  [keybindings_local]="${HOME}/.byobu/keybindings.tmux"
  [datetime_repo]="${REPO_ROOT}/byobu/datetime.tmux"
  [datetime_local]="${HOME}/.byobu/datetime.tmux"
  [statusrc_repo]="${REPO_ROOT}/byobu/statusrc"
  [statusrc_local]="${HOME}/.byobu/statusrc"
)

# Vim paths
declare -gA VIM_PATHS=(
  [vimrc_repo]="${REPO_ROOT}/vim/vimrc"
  [vimrc_local]="${HOME}/.vimrc"
  [colors_repo]="${REPO_ROOT}/vim/colors/sublimemonokai.vim"
  [colors_local]="${HOME}/.vim/colors/sublimemonokai.vim"
)

# Platform detection (must precede OS-conditional path declarations)
if [[ "$OSTYPE" == "darwin"* ]]; then
  export IS_MACOS=true
  export IS_LINUX=false
elif [[ "$OSTYPE" == "linux"* ]]; then
  export IS_MACOS=false
  export IS_LINUX=true
else
  export IS_MACOS=false
  export IS_LINUX=false
fi

# Fonts paths
if [[ "$IS_MACOS" == "true" ]]; then
  declare -gA FONTS_PATHS=(
    [repo]="${REPO_ROOT}/fonts"
    [local]="/Library/Fonts"
  )
else
  declare -gA FONTS_PATHS=(
    [repo]="${REPO_ROOT}/fonts"
    [local]="${HOME}/.local/share/fonts"
  )
fi
