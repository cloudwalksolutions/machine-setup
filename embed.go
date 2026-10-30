// Package tars embeds the repo's dotfiles into the binary so a brew-installed
// tars can provision a machine with no clone. The embed directive can only reach
// files under the module root, which is why the module lives at the repo root.
package tars

import "embed"

//go:embed all:nvim all:zsh all:byobu all:vim all:fonts all:terminal all:claude all:pi all:agents monokai.lua
var Dotfiles embed.FS
