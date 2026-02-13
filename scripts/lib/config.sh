#!/bin/zsh

# Configuration constants for machine-setup scripts
# Defines all file paths and component mappings

# Repo root directory (auto-detected from scripts/lib/)
# scripts/lib/ -> scripts/ -> repo root
REPO_ROOT="$(cd "$(dirname "${(%):-%x}")/../.." && pwd)"

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

# Fonts paths
declare -gA FONTS_PATHS=(
  [repo]="${REPO_ROOT}/fonts"
  [local]="/Library/Fonts"
)

# Platform detection
if [[ "$OSTYPE" == "darwin"* ]]; then
  export IS_MACOS=true
else
  export IS_MACOS=false
fi
