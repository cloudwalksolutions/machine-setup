package pkg

import (
	"github.com/cloudwalk/machine-setup/internal/pkg/apt"
	"github.com/cloudwalk/machine-setup/internal/pkg/brew"
)

// DevToolRegistry owns the list of installables the CLI knows about. It is the
// single source of truth for "what dev tools are available". Composition into
// the registry is one-way (Add/AddAll only) so consumers can rely on the list
// being append-only after construction.
type DevToolRegistry struct {
	tools []Installable
}

// NewDevToolRegistry returns an empty registry.
func NewDevToolRegistry() *DevToolRegistry {
	return &DevToolRegistry{}
}

// Installables returns the registered installables in declaration order.
func (r *DevToolRegistry) Installables() []Installable {
	return r.tools
}

// Add appends an installable; returns the receiver for chaining.
func (r *DevToolRegistry) Add(t Installable) *DevToolRegistry {
	r.tools = append(r.tools, t)
	return r
}

// AddAll appends a batch of installables in order; returns the receiver.
func (r *DevToolRegistry) AddAll(ts []Installable) *DevToolRegistry {
	r.tools = append(r.tools, ts...)
	return r
}

// Names projects the name of each installable.
func (r *DevToolRegistry) Names() []string {
	names := make([]string, len(r.tools))
	for i, t := range r.tools {
		names[i] = t.Name()
	}
	return names
}

// RegistryFactory assembles a DevToolRegistry wired for a given OS. Platform
// runners are captured at construction, plus an optional set of *extras* — any
// Installable the caller wants appended to every supported-OS registry (e.g.
// the RVM curl-pipe installer, which isn't a brew/apt entry).
type RegistryFactory struct {
	brewRun brew.Runner
	aptRun  apt.Runner
	extras  []Installable
}

// NewRegistryFactory captures the platform runners and any cross-platform
// extras. The extras are appended to every recognized-OS registry.
func NewRegistryFactory(brewRun brew.Runner, aptRun apt.Runner, extras ...Installable) RegistryFactory {
	return RegistryFactory{brewRun: brewRun, aptRun: aptRun, extras: extras}
}

// For returns the curated registry for the given OS. Unsupported OS → empty
// (extras are NOT added when no platform is recognized).
func (f RegistryFactory) For(goos string) *DevToolRegistry {
	r := NewDevToolRegistry()
	switch goos {
	case "darwin":
		f.wireDarwin(r)
	case "linux":
		f.wireLinux(r)
	default:
		return r
	}
	r.AddAll(f.extras)
	return r
}

// darwinFormulas is the curated list of brew formulas installed on macOS.
// This is configuration data — adding a tool means adding a name here.
// Tapped formulas (those that need `brew tap` first, like terraform under
// hashicorp/tap) go in darwinTappedFormulas instead.
var darwinFormulas = []string{
	"neovim", "byobu", "fzf", "ripgrep", "bat", "eza",
	"jq", "gh", "go", "node", "python",
	"yarn", "n",
	"rustup", "ghcup",
	"lazygit", "lazydocker", "k9s", "k3d",
	"ruby", "ansible", "golangci-lint",
}

// darwinTappedFormulas pairs each name with its required tap. The TappedFormula
// installer runs `brew tap <tap>` and then `brew install <tap>/<name>`.
var darwinTappedFormulas = map[string]string{
	"terraform": "hashicorp/tap",
}

func (f RegistryFactory) wireDarwin(r *DevToolRegistry) {
	builder := brew.NewBuilder(f.brewRun)
	for _, formula := range builder.Formulas(darwinFormulas...) {
		r.Add(formula)
	}
	for name, tap := range darwinTappedFormulas {
		r.Add(brew.NewTappedFormula(name, tap, f.brewRun))
	}
}

// linuxAptPackages is the curated list of apt packages installed on Linux.
var linuxAptPackages = []string{
	"byobu", "fzf", "ripgrep", "bat",
	"jq", "gh", "go", "node", "python",
}

func (f RegistryFactory) wireLinux(r *DevToolRegistry) {
	r.Add(apt.NeovimAppImage{})
	for _, name := range linuxAptPackages {
		r.Add(apt.NewPackage(name, f.aptRun))
	}
}

