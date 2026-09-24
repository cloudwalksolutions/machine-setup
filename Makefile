# Makefile for tars.
#
# Machine provisioning is the `tars` CLI: tars setup | pull | push | sessions (see README).
# This Makefile is for developing tars + the configs.
#
# Test layers:
#   unit         fast, airgapped logic tests (no external deps, no network)
#   integration  external deps — brew installers + Neovim config tests
#   e2e          the whole thing in Docker, like provisioning a fresh machine
#   test         = unit + integration
#   check        = lint + build + test (the local gate)

SHELL := /bin/bash
.DEFAULT_GOAL := help

CLI := cli

.PHONY: help
help:            ## Show this help
	@echo 'Provisioning: `tars setup | pull | push | sessions` (see README). Dev targets:'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build:           ## Build the tars CLI binary (cli/tars)
	cd $(CLI) && go build -o tars .

.PHONY: sync-assets
sync-assets:     ## Refresh the embedded dotfiles (required after editing them — a drift guard fails CI otherwise)
	rm -rf $(CLI)/internal/assets/tree
	mkdir -p $(CLI)/internal/assets/tree
	cp -R nvim zsh byobu vim fonts terminal claude pi $(CLI)/internal/assets/tree/
	cp monokai.lua $(CLI)/internal/assets/tree/

.PHONY: lint
lint:            ## Lint the CLI (golangci-lint)
	cd $(CLI) && golangci-lint run ./...

.PHONY: unit
unit:            ## Fast, airgapped unit tests
	cd $(CLI) && go test ./...

.PHONY: integration
integration: test-nvim  ## External-dep tests: brew installers + Neovim
	cd $(CLI) && INTEGRATION=1 go test ./internal/pkg/brew/...

.PHONY: test
test: unit integration  ## All tests (unit + integration)

.PHONY: e2e
e2e:             ## Full end-to-end in Docker (fresh-machine simulation)
	docker build -f test/e2e/Dockerfile -t tars-e2e .

.PHONY: check
check: lint build test  ## Local gate: lint, build, then all tests

.PHONY: run
run: build       ## Build and run `tars setup`
	$(CLI)/tars setup

# Recording is staged under /tmp (symlink-resolved: Docker Desktop shares
# /private but not home dirs like ~/Desktop, which macOS privacy blocks).
VHS_STAGE := $(shell cd /tmp && pwd -P)/tars-vhs

.PHONY: demos
demos:           ## Re-record the README demo GIFs (Docker + VHS)
	cd $(CLI) && GOOS=linux go build -o ../vhs/tars-linux .
	docker build -t tars-vhs vhs
	rm -rf $(VHS_STAGE) && mkdir -p $(VHS_STAGE)/vhs
	cp vhs/*.tape $(VHS_STAGE)/ && cp vhs/tars-linux $(VHS_STAGE)/tars
	docker run --rm -v "$(VHS_STAGE):/vhs" -v "$(VHS_STAGE)/tars:/usr/local/bin/tars" tars-vhs demo.tape
	docker run --rm -v "$(VHS_STAGE):/vhs" -v "$(VHS_STAGE)/tars:/usr/local/bin/tars" tars-vhs sessions.tape
	cp $(VHS_STAGE)/vhs/*.gif vhs/

.PHONY: test-nvim
test-nvim:       ## Neovim smoke tests (headless)
	nvim -u nvim/init.lua --headless -l nvim/tests/smoke_test.lua

.PHONY: health-nvim
health-nvim:     ## Neovim health checks
	nvim --headless -c "checkhealth nvim_config" -c "quit"
