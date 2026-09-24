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
				Description: Describe(t.Name()),
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

// darwinFormulas is the curated list of brew formulas installed on macOS.
// This is configuration data — adding a tool means adding a name here.
// Tapped formulas (those that need `brew tap` first, like terraform under
// hashicorp/tap) go in darwinTappedFormulas instead.
var darwinFormulas = []string{
	// editors & terminal, most used first
	"neovim", "byobu", "gh", "lazygit", "jq", "bat", "eza", "k9s", "lazydocker", "k3d", "golangci-lint", "fzf", "ripgrep",
	// languages & runtimes
	"go", "node", "python", "ruby", "rustup", "ghcup", "yarn", "n",
	// devops
	"ansible",
}

// descriptions is the one-line blurb the install picker shows next to each tool.
var descriptions = map[string]string{
	"neovim":        "Modern Vim: LSP, treesitter, Lua config",
	"byobu":         "tmux sessions with a status bar and F-keys",
	"gh":            "GitHub from the terminal: PRs, issues, runs",
	"lazygit":       "Keyboard git UI for staging, log, rebase",
	"jq":            "Query and reshape JSON on the command line",
	"bat":           "cat with syntax highlighting and git marks",
	"eza":           "ls with colors, icons and git status",
	"k9s":           "Kubernetes cluster TUI",
	"lazydocker":    "Docker containers and logs TUI",
	"k3d":           "Local k3s Kubernetes clusters in Docker",
	"golangci-lint": "Go linter aggregator used by CI",
	"fzf":           "Fuzzy finder for files, history, anything",
	"ripgrep":       "Fast recursive grep (rg)",
	"go":            "Go toolchain",
	"node":          "Node.js runtime",
	"python":        "Python 3 interpreter",
	"ruby":          "Ruby interpreter",
	"rustup":        "Rust toolchain manager (installs stable)",
	"ghcup":         "Haskell toolchain manager (installs GHC)",
	"yarn":          "JavaScript package manager",
	"n":             "Switch Node.js versions",
	"ansible":       "Agentless config management and playbooks",
	"terraform":     "Infrastructure as code (HashiCorp tap)",
	"gcloud-cli":    "Google Cloud SDK and gcloud command",
	"gcloud":        "Google Cloud SDK and gcloud command",
	"claude-code":   "Anthropic's Claude Code agent",
	"gemini-cli":    "Google's Gemini CLI agent",
	"pi":            "pi coding agent",
	"rvm":           "Ruby Version Manager",
}

// Describe returns the picker blurb for a tool; empty for unknown names.
func Describe(name string) string { return descriptions[name] }

// darwinTappedFormulas pairs each name with its required tap. The TappedFormula
// installer runs `brew tap <tap>` and then `brew install <tap>/<name>`.
var darwinTappedFormulas = map[string]string{
	"terraform": "hashicorp/tap",
}

// darwinCasks is the curated list of brew casks installed on macOS (GUI apps and
// vendor bundles distributed as casks rather than core formulas).
var darwinCasks = []string{
	"gcloud-cli",
}

// postInstallSteps finish a toolchain manager's setup so no manual step is left to the user.
var postInstallSteps = map[string][][]string{
	"rustup": {{"rustup", "install", "stable"}, {"rustup", "default", "stable"}},
	"ghcup":  {{"ghcup", "install", "ghc", "recommended"}, {"ghcup", "set", "ghc", "recommended"}},
}

func (f RegistryFactory) wireDarwin(r *DevToolRegistry) {
	builder := brew.NewBuilder(f.brewRun)
	for _, formula := range builder.Formulas(darwinFormulas...) {
		if steps, ok := postInstallSteps[formula.Name()]; ok {
			r.Add(pkg.WithPostInstall(formula, steps, f.aptKit.Cmd))
			continue
		}
		r.Add(formula)
	}
	for name, tap := range darwinTappedFormulas {
		r.Add(brew.NewTappedFormula(name, tap, f.brewRun))
	}
	for _, cask := range builder.Casks(darwinCasks...) {
		r.Add(cask)
	}
}

// linuxAptPackages is the curated list of apt packages installed on Linux.
// gh is NOT here — Ubuntu's archives don't carry it; see GitHubCLI below.
var linuxAptPackages = []string{
	"jq", "bat", "fzf", "ripgrep",
	"go", "node", "python",
}

func (f RegistryFactory) wireLinux(r *DevToolRegistry) {
	r.Add(apt.NeovimTarball{Fetch: f.aptKit.Fetch, Home: f.aptKit.Home, Cmd: f.aptKit.Cmd})
	r.Add(apt.NewPackage("byobu", f.aptKit.Apt, f.aptKit.Cmd))
	r.Add(apt.NewGitHubCLI(f.aptKit.Cmd))
	for _, name := range linuxAptPackages {
		r.Add(apt.NewPackage(name, f.aptKit.Apt, f.aptKit.Cmd))
	}
	r.Add(apt.NewGCloudCLI(f.aptKit.Cmd))
}
