// Package paths defines the repo→local file mappings for every component.
package paths

import (
	"path/filepath"
	"runtime"
)

// Paths holds all repo→local file mappings for each component.
type Paths struct {
	Nvim     NvimPaths
	Zsh      ZshPaths
	Byobu    ByobuPaths
	Vim      VimPaths
	Fonts    FontsPaths
	Terminal TerminalPaths
	Profiles ProfilesPaths
}

// ProfilesPaths locates the rendered per-profile files and the gitconfig they hook into.
type ProfilesPaths struct {
	Dir       string
	Gitconfig string
}

// TerminalPaths locates the font string applied to iTerm2's default profile.
type TerminalPaths struct {
	FontRepo    string
	ProfileRepo string
}

// FontsPaths holds the font repo→local mapping. Local destination
// is OS-dependent: /Library/Fonts on darwin, ~/.local/share/fonts elsewhere.
type FontsPaths struct {
	Repo  string
	Local string
}

// VimPaths holds the vim repo→local mappings.
type VimPaths struct {
	VimrcRepo   string
	VimrcLocal  string
	ColorsRepo  string
	ColorsLocal string
}

// NvimPaths holds the neovim repo→local mappings.
type NvimPaths struct {
	Repo         string
	Local        string
	MonokaiRepo  string
	MonokaiLocal string
}

// ZshPaths holds the zsh repo→local mappings, plus the two create-only
// templates setup seeds into ~/.zshrc_secret and ~/.zprofile_local.
type ZshPaths struct {
	ZshrcRepo            string
	ZshrcLocal           string
	AliasesRepo          string
	AliasesLocal         string
	FuncsRepo            string
	FuncsLocal           string
	ProfileRepo          string
	ProfileLocal         string
	ProfileLocalTemplate string
	ProfileLocalOverride string
	SecretTemplate       string
	SecretLocal          string
}

// ByobuPaths holds the byobu repo→local mappings.
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
	ColorRepo        string
	ColorLocal       string
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

	// zsh login shells read ~/.zprofile and never ~/.profile, so on macOS
	// (zsh by default) the login profile must target ~/.zprofile. Other OSes
	// keep ~/.profile.
	profileLocal := filepath.Join(home, ".profile")
	if goos == "darwin" {
		profileLocal = filepath.Join(home, ".zprofile")
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
			ZshrcRepo:            filepath.Join(repoRoot, "zsh", "zshrc"),
			ZshrcLocal:           filepath.Join(home, ".zshrc"),
			AliasesRepo:          filepath.Join(repoRoot, "zsh", "zshrc_aliases"),
			AliasesLocal:         filepath.Join(home, ".zshrc_aliases"),
			FuncsRepo:            filepath.Join(repoRoot, "zsh", "zshrc_funcs"),
			FuncsLocal:           filepath.Join(home, ".zshrc_funcs"),
			ProfileRepo:          filepath.Join(repoRoot, "zsh", "profile"),
			ProfileLocal:         profileLocal,
			ProfileLocalTemplate: filepath.Join(repoRoot, "zsh", "zprofile_local.template"),
			ProfileLocalOverride: filepath.Join(home, ".zprofile_local"),
			SecretTemplate:       filepath.Join(repoRoot, "zsh", "zshrc_secret.template"),
			SecretLocal:          filepath.Join(home, ".zshrc_secret"),
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
			ColorRepo:        filepath.Join(repoRoot, "byobu", "color.tmux"),
			ColorLocal:       filepath.Join(home, ".byobu", "color.tmux"),
		},
		Vim: VimPaths{
			VimrcRepo:   filepath.Join(repoRoot, "vim", "vimrc"),
			VimrcLocal:  filepath.Join(home, ".vimrc"),
			ColorsRepo:  filepath.Join(repoRoot, "vim", "colors", "sublimemonokai.vim"),
			ColorsLocal: filepath.Join(home, ".vim", "colors", "sublimemonokai.vim"),
		},
		Terminal: TerminalPaths{
			FontRepo:    filepath.Join(repoRoot, "terminal", "font"),
			ProfileRepo: filepath.Join(repoRoot, "terminal", "tars.terminal"),
		},
		Profiles: ProfilesPaths{
			Dir:       filepath.Join(home, ".config", "tars", "profiles"),
			Gitconfig: filepath.Join(home, ".gitconfig"),
		},
	}
}
