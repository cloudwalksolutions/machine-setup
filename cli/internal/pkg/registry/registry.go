package registry

import (
	"sync"

	"tars/internal/pkg"
	"tars/internal/pkg/apt"
	"tars/internal/pkg/brew"
)

// DevToolRegistry owns the list of installables the CLI knows about. It is the
// single source of truth for "what dev tools are available". Composition into
// the registry is one-way (Add/AddAll only) so consumers can rely on the list
// being append-only after construction.
type DevToolRegistry struct {
	tools []pkg.Installable
}

// NewDevToolRegistry returns an empty registry.
func NewDevToolRegistry() *DevToolRegistry {
	return &DevToolRegistry{}
}

// Installables returns the registered installables in declaration order.
func (r *DevToolRegistry) Installables() []pkg.Installable {
	return r.tools
}

// Add appends an installable; returns the receiver for chaining.
func (r *DevToolRegistry) Add(t pkg.Installable) *DevToolRegistry {
	r.tools = append(r.tools, t)
	return r
}

// AddAll appends a batch of installables in order; returns the receiver.
func (r *DevToolRegistry) AddAll(ts []pkg.Installable) *DevToolRegistry {
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

// Catalog projects each installable into what the picker shows, in order; statuses are
// queried concurrently because each brew/npm query takes a noticeable fraction of a second.
func (r *DevToolRegistry) Catalog() []pkg.ToolInfo {
	infos := make([]pkg.ToolInfo, len(r.tools))
	var wg sync.WaitGroup
	for i, t := range r.tools {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, version, _ := t.Status()
			infos[i] = pkg.ToolInfo{
				Name:        t.Name(),
				Description: t.Description(),
				Installed:   status != pkg.StatusNotInstalled,
				Version:     version,
			}
		}()
	}
	wg.Wait()
	return infos
}

// RegistryFactory assembles a DevToolRegistry wired for a given OS. Platform
// runners are captured at construction, plus an optional set of *extras* — any
// Installable the caller wants appended to every supported-OS registry (e.g.
// the RVM curl-pipe installer, which isn't a brew/apt entry).
type RegistryFactory struct {
	brewRun brew.Runner
	aptKit  apt.Kit
	probe   pkg.PathProbe
	extras  []pkg.Installable
}

// NewRegistryFactory captures the platform runners, the PATH probe used when a
// package manager does not know a tool, and any cross-platform extras.
func NewRegistryFactory(brewRun brew.Runner, aptKit apt.Kit, probe pkg.PathProbe, extras ...pkg.Installable) RegistryFactory {
	return RegistryFactory{brewRun: brewRun, aptKit: aptKit, probe: probe, extras: extras}
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

// tool is one curated entry: the package name, the blurb the picker shows next to it,
// and the binary to look for on PATH when it differs from the name.
type tool struct {
	name, description, bin string
}

// darwinFormulas is the curated list of brew formulas installed on macOS, most used
// first within each group. Tapped formulas go in darwinTappedFormulas instead.
var darwinFormulas = []tool{
	{name: "neovim", description: "Modern Vim: LSP, treesitter, Lua config", bin: "nvim"},
	{name: "byobu", description: "tmux sessions with a status bar and F-keys"},
	{name: "gh", description: "GitHub from the terminal: PRs, issues, runs"},
	{name: "lazygit", description: "Keyboard git UI for staging, log, rebase"},
	{name: "jq", description: "Query and reshape JSON on the command line"},
	{name: "bat", description: "cat with syntax highlighting and git marks"},
	{name: "eza", description: "ls with colors, icons and git status"},
	{name: "k9s", description: "Kubernetes cluster TUI"},
	{name: "lazydocker", description: "Docker containers and logs TUI"},
	{name: "k3d", description: "Local k3s Kubernetes clusters in Docker"},
	{name: "golangci-lint", description: "Go linter aggregator used by CI"},
	{name: "fzf", description: "Fuzzy finder for files, history, anything"},
	{name: "ripgrep", description: "Fast recursive grep (rg)", bin: "rg"},
	{name: "go", description: "Go toolchain"},
	{name: "node", description: "Node.js runtime"},
	{name: "python", description: "Python 3 interpreter", bin: "python3"},
	{name: "ruby", description: "Ruby interpreter"},
	{name: "rustup", description: "Rust toolchain manager (installs stable)"},
	{name: "ghcup", description: "Haskell toolchain manager (installs GHC)"},
	{name: "yarn", description: "JavaScript package manager"},
	{name: "n", description: "Switch Node.js versions"},
	{name: "ansible", description: "Agentless config management and playbooks"},
}

// darwinTappedFormulas need `brew tap <tap>` first; installed as <tap>/<name>.
var darwinTappedFormulas = []struct {
	tool
	tap string
}{
	{tool{name: "terraform", description: "Infrastructure as code (HashiCorp tap)"}, "hashicorp/tap"},
}

// darwinCasks is the curated list of brew casks installed on macOS.
var darwinCasks = []tool{
	{name: "gcloud-cli", description: "Google Cloud SDK and gcloud command", bin: "gcloud"},
}

// postInstallSteps finish a toolchain manager's setup so no manual step is left to the user.
var postInstallSteps = map[string][][]string{
	"rustup": {{"rustup", "install", "stable"}, {"rustup", "default", "stable"}},
	"ghcup":  {{"ghcup", "install", "ghc", "recommended"}, {"ghcup", "set", "ghc", "recommended"}},
}

func (f RegistryFactory) wireDarwin(r *DevToolRegistry) {
	for _, t := range darwinFormulas {
		formula := brew.NewFormula(t.name, t.description, f.brewRun)
		formula.Binary, formula.Probe = t.bin, f.probe
		if steps, ok := postInstallSteps[t.name]; ok {
			r.Add(pkg.WithPostInstall(formula, steps, f.aptKit.Cmd))
			continue
		}
		r.Add(formula)
	}
	for _, t := range darwinTappedFormulas {
		tapped := brew.NewTappedFormula(t.name, t.description, t.tap, f.brewRun)
		tapped.Probe = f.probe
		r.Add(tapped)
	}
	for _, t := range darwinCasks {
		cask := brew.NewCask(t.name, t.description, f.brewRun)
		cask.Binary, cask.Probe = t.bin, f.probe
		r.Add(cask)
	}
}

// linuxAptPackages is the curated list of apt packages installed on Linux.
// gh is NOT here — Ubuntu's archives don't carry it; see GitHubCLI.
var linuxAptPackages = []tool{
	{name: "jq", description: "Query and reshape JSON on the command line"},
	{name: "bat", description: "cat with syntax highlighting and git marks"},
	{name: "fzf", description: "Fuzzy finder for files, history, anything"},
	{name: "ripgrep", description: "Fast recursive grep (rg)"},
	{name: "go", description: "Go toolchain"},
	{name: "node", description: "Node.js runtime"},
	{name: "python", description: "Python 3 interpreter"},
}

func (f RegistryFactory) wireLinux(r *DevToolRegistry) {
	r.Add(apt.NeovimTarball{Fetch: f.aptKit.Fetch, Home: f.aptKit.Home, Cmd: f.aptKit.Cmd})
	r.Add(apt.NewPackage("byobu", "tmux sessions with a status bar and F-keys", f.aptKit.Apt, f.aptKit.Cmd))
	r.Add(apt.NewGitHubCLI(f.aptKit.Cmd))
	for _, t := range linuxAptPackages {
		r.Add(apt.NewPackage(t.name, t.description, f.aptKit.Apt, f.aptKit.Cmd))
	}
	r.Add(apt.NewGCloudCLI(f.aptKit.Cmd))
}
