package pkg

import (
	"fmt"
	"io"
	"runtime"

	"github.com/cloudwalk/machine-setup/internal/config"
	"github.com/cloudwalk/machine-setup/internal/pkg/brew"
)

// Manager abstracts package manager operations.
// Implement this interface to add apt, choco, scoop, etc. alongside brew.
type Manager interface {
	Install(name string) error
	Uninstall(name string) error
	IsInstalled(name string) (bool, error)
	Update(name string) error
}

// DevTools is the curated CloudWalk default package list.
var DevTools = []config.Package{
	{Name: "neovim", Manager: "brew"},
	{Name: "starship", Manager: "brew"},
	{Name: "byobu", Manager: "brew"},
	{Name: "fzf", Manager: "brew"},
	{Name: "ripgrep", Manager: "brew"},
	{Name: "bat", Manager: "brew"},
	{Name: "eza", Manager: "brew"},
	{Name: "jq", Manager: "brew"},
	{Name: "gh", Manager: "brew"},
	{Name: "go", Manager: "brew"},
	{Name: "node", Manager: "brew"},
	{Name: "python", Manager: "brew"},
}

// DevToolNames returns the names of all DevTools.
func DevToolNames() []string {
	names := make([]string, len(DevTools))
	for i, p := range DevTools {
		names[i] = p.Name
	}
	return names
}

// NamesToPackages converts a slice of names back to []config.Package,
// matching against DevTools for the manager field.
func NamesToPackages(names []string) []config.Package {
	lookup := make(map[string]config.Package, len(DevTools))
	for _, p := range DevTools {
		lookup[p.Name] = p
	}
	pkgs := make([]config.Package, 0, len(names))
	for _, name := range names {
		if p, ok := lookup[name]; ok {
			pkgs = append(pkgs, p)
		} else {
			pkgs = append(pkgs, config.Package{Name: name, Manager: "brew"})
		}
	}
	return pkgs
}

// NewManager auto-detects and returns the right Manager for the current OS.
func NewManager(stdout, stderr io.Writer) (Manager, error) {
	switch runtime.GOOS {
	case "darwin":
		return brew.New(stdout, stderr), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
