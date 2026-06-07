// Package paths ports the path tables from scripts/lib/config.sh.
package paths

import (
	"path/filepath"
	"runtime"
)

// Paths holds all repo→local file mappings for each component.
type Paths struct {
	Nvim  NvimPaths
	Zsh   ZshPaths
	Byobu ByobuPaths
	Vim   VimPaths
	Fonts FontsPaths
}

// FontsPaths mirrors FONTS_PATHS in scripts/lib/config.sh. Local destination
// is OS-dependent: /Library/Fonts on darwin, ~/.local/share/fonts elsewhere.
type FontsPaths struct {
	Repo  string
	Local string
}

// VimPaths mirrors VIM_PATHS in scripts/lib/config.sh.
type VimPaths struct {
	VimrcRepo   string
	VimrcLocal  string
	ColorsRepo  string
	ColorsLocal string
}

// NvimPaths mirrors NVIM_PATHS in scripts/lib/config.sh.
type NvimPaths struct {
	Repo         string
	Local        string
	MonokaiRepo  string
	MonokaiLocal string
}

// ZshPaths mirrors ZSH_PATHS in scripts/lib/config.sh, plus the secret template
// which init.sh handles separately.
type ZshPaths struct {
	ZshrcRepo      string
	ZshrcLocal     string
	AliasesRepo    string
	AliasesLocal   string
	FuncsRepo      string
	FuncsLocal     string
	ProfileRepo    string
	ProfileLocal   string
	SecretTemplate string
	SecretLocal    string
}

// ByobuPaths mirrors BYOBU_PATHS in scripts/lib/config.sh.
type ByobuPaths struct {
	BinRepo          string
	BinLocal         string
	TmuxConfRepo     string
	TmuxConfLocal    string
	KeybindingsRepo  string
	KeybindingsLocal string
	DatetimeRepo     string
	DatetimeLocal    string
	StatusrcRepo     string
	StatusrcLocal    string
}

// For builds a Paths bundle rooted at the given repo and home directories,
// using the current OS for OS-dependent fields (fonts).
func For(repoRoot, home string) Paths {
	return ForOS(repoRoot, home, runtime.GOOS)
}

// ForOS is the OS-explicit form of For, used by tests.
func ForOS(repoRoot, home, goos string) Paths {
	fontsLocal := filepath.Join(home, ".local", "share", "fonts")
	if goos == "darwin" {
		fontsLocal = "/Library/Fonts"
	}
	return Paths{
		Fonts: FontsPaths{
			Repo:  filepath.Join(repoRoot, "fonts"),
			Local: fontsLocal,
		},
		Nvim: NvimPaths{
			Repo:         filepath.Join(repoRoot, "nvim"),
			Local:        filepath.Join(home, ".config", "nvim"),
			MonokaiRepo:  filepath.Join(repoRoot, "monokai.lua"),
			MonokaiLocal: filepath.Join(home, ".local", "share", "nvim", "site", "pack", "packer", "start", "monokai.nvim", "lua"),
		},
		Zsh: ZshPaths{
			ZshrcRepo:      filepath.Join(repoRoot, "zsh", "zshrc"),
			ZshrcLocal:     filepath.Join(home, ".zshrc"),
			AliasesRepo:    filepath.Join(repoRoot, "zsh", "zshrc_aliases"),
			AliasesLocal:   filepath.Join(home, ".zshrc_aliases"),
			FuncsRepo:      filepath.Join(repoRoot, "zsh", "zshrc_funcs"),
			FuncsLocal:     filepath.Join(home, ".zshrc_funcs"),
			ProfileRepo:    filepath.Join(repoRoot, "zsh", "profile"),
			ProfileLocal:   filepath.Join(home, ".profile"),
			SecretTemplate: filepath.Join(repoRoot, "zsh", "zshrc_secret.template"),
			SecretLocal:    filepath.Join(home, ".zshrc_secret"),
		},
		Byobu: ByobuPaths{
			BinRepo:          filepath.Join(repoRoot, "byobu", "bin"),
			BinLocal:         filepath.Join(home, ".byobu", "bin"),
			TmuxConfRepo:     filepath.Join(repoRoot, "byobu", ".tmux.conf"),
			TmuxConfLocal:    filepath.Join(home, ".byobu", ".tmux.conf"),
			KeybindingsRepo:  filepath.Join(repoRoot, "byobu", "keybindings.tmux"),
			KeybindingsLocal: filepath.Join(home, ".byobu", "keybindings.tmux"),
			DatetimeRepo:     filepath.Join(repoRoot, "byobu", "datetime.tmux"),
			DatetimeLocal:    filepath.Join(home, ".byobu", "datetime.tmux"),
			StatusrcRepo:     filepath.Join(repoRoot, "byobu", "statusrc"),
			StatusrcLocal:    filepath.Join(home, ".byobu", "statusrc"),
		},
		Vim: VimPaths{
			VimrcRepo:   filepath.Join(repoRoot, "vim", "vimrc"),
			VimrcLocal:  filepath.Join(home, ".vimrc"),
			ColorsRepo:  filepath.Join(repoRoot, "vim", "colors", "sublimemonokai.vim"),
			ColorsLocal: filepath.Join(home, ".vim", "colors", "sublimemonokai.vim"),
		},
	}
}
