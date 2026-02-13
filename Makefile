# Makefile for machine-setup repository
# Provides user-friendly interface for configuration management

SHELL := /bin/zsh

.DEFAULT_GOAL := help

.PHONY: help
help:                           ## Show this help message
	@echo '🔧 Machine Setup - Configuration Management'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Main targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: init
init:                           ## Fresh machine setup (brew install + pull configs)
	@echo '🚀 Initializing machine setup...'
	@./scripts/init.sh

.PHONY: pull
pull:                           ## Copy configs from repo → local machine (with backup)
	@echo '⬇️  Pulling configurations...'
	@./scripts/pull.sh

.PHONY: push
push:                           ## Copy configs from local machine → repo (with git commit)
	@echo '⬆️  Pushing configurations...'
	@./scripts/push.sh

.PHONY: backup
backup:                         ## Create manual backup of all local configs
	@echo '💾 Creating manual backup...'
	@. ./scripts/lib/common.sh && backup_all

.PHONY: backups-list
backups-list:                   ## List all backup versions
	@echo '📋 Listing all backups...'
	@echo ''
	@. ./scripts/lib/common.sh && list_all_backups

.PHONY: migrate
migrate:                        ## Migrate old dot-based filenames to underscore naming
	@echo '🔄 Migrating to new naming convention...'
	@./scripts/migrate.sh

.PHONY: pull-nvim
pull-nvim:                      ## Pull only Neovim config
	@echo '⬇️  Pulling Neovim config...'
	@./scripts/components/nvim.sh pull

.PHONY: pull-zsh
pull-zsh:                       ## Pull only Zsh config
	@echo '⬇️  Pulling Zsh config...'
	@./scripts/components/zsh.sh pull

.PHONY: pull-byobu
pull-byobu:                     ## Pull only Byobu config
	@echo '⬇️  Pulling Byobu config...'
	@./scripts/components/byobu.sh pull

.PHONY: pull-vim
pull-vim:                       ## Pull only Vim config
	@echo '⬇️  Pulling Vim config...'
	@./scripts/components/vim.sh pull

.PHONY: pull-fonts
pull-fonts:                     ## Pull only fonts
	@echo '⬇️  Installing fonts...'
	@./scripts/components/fonts.sh pull

.PHONY: push-nvim
push-nvim:                      ## Push only Neovim config
	@echo '⬆️  Pushing Neovim config...'
	@./scripts/components/nvim.sh push

.PHONY: push-zsh
push-zsh:                       ## Push only Zsh config
	@echo '⬆️  Pushing Zsh config...'
	@./scripts/components/zsh.sh push

.PHONY: push-byobu
push-byobu:                     ## Push only Byobu config
	@echo '⬆️  Pushing Byobu config...'
	@./scripts/components/byobu.sh push

.PHONY: push-vim
push-vim:                       ## Push only Vim config
	@echo '⬆️  Pushing Vim config...'
	@./scripts/components/vim.sh push
