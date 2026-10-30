// The embed directive cannot reach above the module root, which is why the module lives at the repo root.
package tars

import "embed"

//go:embed all:nvim all:zsh all:byobu all:vim all:fonts all:terminal all:claude all:pi all:agents monokai.lua
var Dotfiles embed.FS
