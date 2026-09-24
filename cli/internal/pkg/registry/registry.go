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
	extras  []pkg.Installable
}

// NewRegistryFactory captures the platform runners and any cross-platform
// extras. The extras are appended to every recognized-OS registry.
func NewRegistryFactory(brewRun brew.Runner, aptKit apt.Kit, extras ...pkg.Installable) RegistryFactory {
	return RegistryFactory{brewRun: brewRun, aptKit: aptKit, extras: extras}
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

// tool is one curated entry: the package name and the blurb the picker shows next to it.
type tool struct {
	name, description string
}

// darwinFormulas is the curated list of brew formulas installed on macOS, most used
// first within each group. Tapped formulas go in darwinTappedFormulas instead.
var darwinFormulas = []tool{
	{"neovim", "Modern Vim: LSP, treesitter, Lua config"},
	{"byobu", "tmux sessions with a status bar and F-keys"},
	{"gh", "GitHub from the terminal: PRs, issues, runs"},
	{"lazygit", "Keyboard git UI for staging, log, rebase"},
	{"jq", "Query and reshape JSON on the command line"},
	{"bat", "cat with syntax highlighting and git marks"},
	{"eza", "ls with colors, icons and git status"},
	{"k9s", "Kubernetes cluster TUI"},
	{"lazydocker", "Docker containers and logs TUI"},
	{"k3d", "Local k3s Kubernetes clusters in Docker"},
	{"golangci-lint", "Go linter aggregator used by CI"},
	{"fzf", "Fuzzy finder for files, history, anything"},
	{"ripgrep", "Fast recursive grep (rg)"},
	{"go", "Go toolchain"},
	{"node", "Node.js runtime"},
	{"python", "Python 3 interpreter"},
	{"ruby", "Ruby interpreter"},
	{"rustup", "Rust toolchain manager (installs stable)"},
	{"ghcup", "Haskell toolchain manager (installs GHC)"},
	{"yarn", "JavaScript package manager"},
	{"n", "Switch Node.js versions"},
	{"ansible", "Agentless config management and playbooks"},
}

// darwinTappedFormulas need `brew tap <tap>` first; installed as <tap>/<name>.
var darwinTappedFormulas = []struct {
	tool
	tap string
}{
	{tool{"terraform", "Infrastructure as code (HashiCorp tap)"}, "hashicorp/tap"},
}

// darwinCasks is the curated list of brew casks installed on macOS.
var darwinCasks = []tool{
	{"gcloud-cli", "Google Cloud SDK and gcloud command"},
}

// postInstallSteps finish a toolchain manager's setup so no manual step is left to the user.
var postInstallSteps = map[string][][]string{
	"rustup": {{"rustup", "install", "stable"}, {"rustup", "default", "stable"}},
	"ghcup":  {{"ghcup", "install", "ghc", "recommended"}, {"ghcup", "set", "ghc", "recommended"}},
}

func (f RegistryFactory) wireDarwin(r *DevToolRegistry) {
	for _, t := range darwinFormulas {
		formula := brew.NewFormula(t.name, t.description, f.brewRun)
		if steps, ok := postInstallSteps[t.name]; ok {
			r.Add(pkg.WithPostInstall(formula, steps, f.aptKit.Cmd))
			continue
		}
		r.Add(formula)
	}
	for _, t := range darwinTappedFormulas {
		r.Add(brew.NewTappedFormula(t.name, t.description, t.tap, f.brewRun))
	}
	for _, t := range darwinCasks {
		r.Add(brew.NewCask(t.name, t.description, f.brewRun))
	}
}

// linuxAptPackages is the curated list of apt packages installed on Linux.
// gh is NOT here — Ubuntu's archives don't carry it; see GitHubCLI.
var linuxAptPackages = []tool{
	{"jq", "Query and reshape JSON on the command line"},
	{"bat", "cat with syntax highlighting and git marks"},
	{"fzf", "Fuzzy finder for files, history, anything"},
	{"ripgrep", "Fast recursive grep (rg)"},
	{"go", "Go toolchain"},
	{"node", "Node.js runtime"},
	{"python", "Python 3 interpreter"},
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
