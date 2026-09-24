// Package apt provides polymorphic Installable implementations for the
// Debian/Ubuntu side of the CLI's curated install list. Each installable
// type (Package, NeovimAppImage) is its own kind that knows how to install
// itself — no dispatcher map, no type switches.
package apt

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"tars/internal/pkg"
)

// Runner runs an apt subcommand. Production wiring shells out to
// `sudo apt …`; tests inject a recorder.
type Runner func(args []string, stdout, stderr io.Writer) error

// SudoAptArgs builds the argv for an apt subcommand: non-interactive sudo
// (-n fails fast instead of hanging on a password prompt with no TTY),
// DEBIAN_FRONTEND=noninteractive, and apt-get (apt's CLI is not script-stable).
func SudoAptArgs(args []string) []string {
	return append([]string{"sudo", "-n", "DEBIAN_FRONTEND=noninteractive", "apt-get"}, args...)
}

// DefaultRunner returns the production Runner.
func DefaultRunner() Runner {
	return func(args []string, stdout, stderr io.Writer) error {
		argv := SudoAptArgs(args)
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
}

// CmdRunner executes one argv (sudo/sh steps included) — the seam for
// multi-step installers so specs record steps instead of exec-ing.
type CmdRunner func(argv []string, stdout, stderr io.Writer) error

// DefaultCmdRunner returns the production CmdRunner.
func DefaultCmdRunner() CmdRunner {
	return func(argv []string, stdout, stderr io.Writer) error {
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
}

// Kit bundles every side-effecting dependency of the apt installables so the
// registry can wire one value and specs can inject spies for all of them.
type Kit struct {
	Apt   Runner
	Cmd   CmdRunner
	Fetch Fetcher
	Home  string
}

// DefaultKit returns the production wiring.
func DefaultKit() Kit {
	home, _ := os.UserHomeDir()
	return Kit{
		Apt:   DefaultRunner(),
		Cmd:   DefaultCmdRunner(),
		Fetch: DefaultFetcher(),
		Home:  home,
	}
}

// aptNames maps brew-style names to their apt equivalents.
var aptNames = map[string]string{
	"go":     "golang",
	"node":   "nodejs",
	"python": "python3",
}

// Package is an apt-installable package referenced by its brew-style name.
// Install resolves the name to the apt package and runs `apt install -y`.
type Package struct {
	name        string
	description string
	run         Runner
	query       CmdRunner
}

// NewPackage binds a package to the apt runner and, optionally, a CmdRunner for dpkg-query.
func NewPackage(name, description string, run Runner, query ...CmdRunner) Package {
	p := Package{name: name, description: description, run: run}
	if len(query) > 0 {
		p.query = query[0]
	}
	return p
}

// Name returns the brew-style name (unresolved). This is what the user sees.
func (p Package) Name() string { return p.name }

// Description returns the picker blurb.
func (p Package) Description() string { return p.description }

func (p Package) resolved() string {
	if mapped, ok := aptNames[p.name]; ok {
		return mapped
	}
	return p.name
}

// Install runs `apt install -y <resolved-name>`.
func (p Package) Install(stdout, stderr io.Writer) error {
	return p.run([]string{"install", "-y", p.resolved()}, stdout, stderr)
}

// Status reads the installed version via dpkg-query; a non-zero exit means not installed.
func (p Package) Status() (pkg.InstallStatus, string, error) {
	return dpkgVersion(p.query, p.resolved())
}

// dpkgVersion asks dpkg for the installed version of one package.
func dpkgVersion(run CmdRunner, name string) (pkg.InstallStatus, string, error) {
	if run == nil {
		run = DefaultCmdRunner()
	}
	var out bytes.Buffer
	if err := run([]string{"dpkg-query", "-W", "-f=${Version}", name}, &out, io.Discard); err != nil {
		return pkg.StatusNotInstalled, "", nil
	}
	version := strings.TrimSpace(out.String())
	if version == "" {
		return pkg.StatusNotInstalled, "", nil
	}
	return pkg.StatusUpToDate, version, nil
}

// Fetcher performs an HTTP GET and returns the body — the seam for downloads.
type Fetcher func(url string) (io.ReadCloser, error)

// DefaultFetcher returns the production Fetcher (non-200 is an error).
func DefaultFetcher() Fetcher {
	return func(url string) (io.ReadCloser, error) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
}

// NeovimTarball installs Neovim from the upstream release tarball, extracted
// to <Home>/.local/nvim with a symlink at <Home>/.local/bin/nvim. Used on
// Linux where the apt package is outdated; unlike the AppImage it needs no FUSE.
type NeovimTarball struct {
	Fetch Fetcher
	Home  string
	Arch  string // "amd64" | "arm64"; empty means runtime.GOARCH
	Cmd   CmdRunner
}

// Name reports "neovim" to match its brew counterpart for the form display.
func (NeovimTarball) Name() string { return "neovim" }

// Description returns the picker blurb.
func (NeovimTarball) Description() string { return "Modern Vim: LSP, treesitter, Lua config" }

// Status reports the version the extracted binary prints, once ~/.local/nvim exists.
func (n NeovimTarball) Status() (pkg.InstallStatus, string, error) {
	destRoot := filepath.Join(n.Home, ".local", "nvim")
	if _, err := os.Stat(destRoot); err != nil {
		return pkg.StatusNotInstalled, "", nil
	}
	run := n.Cmd
	if run == nil {
		run = DefaultCmdRunner()
	}
	var out bytes.Buffer
	if err := run([]string{filepath.Join(destRoot, "bin", "nvim"), "--version"}, &out, io.Discard); err != nil {
		return pkg.StatusUpToDate, "", nil
	}
	fields := strings.Fields(out.String())
	if len(fields) < 2 {
		return pkg.StatusUpToDate, "", nil
	}
	return pkg.StatusUpToDate, fields[1], nil
}

// Install downloads and extracts the tarball, then links the binary onto PATH.
func (n NeovimTarball) Install(stdout, stderr io.Writer) error {
	arch := n.Arch
	if arch == "" {
		arch = runtime.GOARCH
	}
	if arch == "amd64" {
		arch = "x86_64"
	}

	url := fmt.Sprintf("https://github.com/neovim/neovim/releases/download/v0.11.6/nvim-linux-%s.tar.gz", arch)
	destRoot := filepath.Join(n.Home, ".local", "nvim")

	fmt.Fprintf(stdout, "Downloading Neovim to %s...\n", destRoot)
	body, err := n.Fetch(url)
	if err != nil {
		return fmt.Errorf("downloading neovim: %w", err)
	}
	defer body.Close()

	if err := extractTarGz(body, destRoot); err != nil {
		return fmt.Errorf("extracting neovim: %w", err)
	}

	link := filepath.Join(n.Home, ".local", "bin", "nvim")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(link); err != nil {
		return err
	}
	return os.Symlink(filepath.Join(destRoot, "bin", "nvim"), link)
}

// extractTarGz unpacks a gzipped tarball into destRoot, stripping the single
// top-level directory and honoring tar file modes.
func extractTarGz(r io.Reader, destRoot string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		rel := stripTopDir(hdr.Name)
		if rel == "" {
			continue
		}
		// Reject entries that would escape destRoot (zip-slip).
		dest := filepath.Join(destRoot, rel)
		if !strings.HasPrefix(dest, destRoot+string(filepath.Separator)) {
			return fmt.Errorf("tar entry escapes destination: %q", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode).Perm())
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil { //nolint:gosec // trusted release asset
				_ = out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
			if err := os.Chmod(dest, os.FileMode(hdr.Mode).Perm()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			if err := os.Symlink(hdr.Linkname, dest); err != nil {
				return err
			}
		}
	}
}

// stripTopDir drops the tarball's single top-level directory from a path.
func stripTopDir(name string) string {
	name = strings.TrimPrefix(name, "./")
	if _, rest, found := strings.Cut(name, "/"); found {
		return strings.Trim(rest, "/")
	}
	return ""
}

// GCloudCLI installs the Google Cloud CLI on Debian/Ubuntu by adding Google's
// apt repository (the package is not in the default archives) and installing
// google-cloud-cli. Mirrors the documented steps at
// https://cloud.google.com/sdk/docs/install#deb.
type GCloudCLI struct {
	Run CmdRunner
}

// NewGCloudCLI returns a GCloudCLI bound to a step runner.
func NewGCloudCLI(run CmdRunner) GCloudCLI { return GCloudCLI{Run: run} }

// Name reports "gcloud" to match its brew-cask counterpart for the form display.
func (GCloudCLI) Name() string { return "gcloud" }

// Description returns the picker blurb.
func (GCloudCLI) Description() string { return "Google Cloud SDK and gcloud command" }

// Install adds Google's apt source and key, then installs google-cloud-cli.
func (g GCloudCLI) Install(stdout, stderr io.Writer) error {
	const keyring = "/usr/share/keyrings/cloud.google.gpg"
	steps := [][]string{
		SudoAptArgs([]string{"update"}),
		SudoAptArgs([]string{"install", "-y", "apt-transport-https", "ca-certificates", "gnupg", "curl"}),
		// Fetch Google's signing key and dearmor it into the keyring.
		{"sh", "-c", "curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo -n gpg --dearmor --yes -o " + keyring},
		// Register the cloud-sdk apt source, pinned to the keyring and this arch.
		{"sh", "-c", "echo \"deb [arch=$(dpkg --print-architecture) signed-by=" + keyring + "] https://packages.cloud.google.com/apt cloud-sdk main\" | sudo -n tee /etc/apt/sources.list.d/google-cloud-sdk.list"},
		SudoAptArgs([]string{"update"}),
		SudoAptArgs([]string{"install", "-y", "google-cloud-cli"}),
	}
	return runSteps("gcloud", steps, g.Run, stdout, stderr)
}

// Status reads the installed google-cloud-cli version via dpkg-query.
func (g GCloudCLI) Status() (pkg.InstallStatus, string, error) {
	return dpkgVersion(g.Run, "google-cloud-cli")
}

// GitHubCLI installs gh on Debian/Ubuntu by adding GitHub's apt repository
// (gh is not in the default archives). Mirrors the documented steps at
// https://github.com/cli/cli/blob/trunk/docs/install_linux.md.
type GitHubCLI struct {
	Run CmdRunner
}

// NewGitHubCLI returns a GitHubCLI bound to a step runner.
func NewGitHubCLI(run CmdRunner) GitHubCLI { return GitHubCLI{Run: run} }

// Name reports "gh" to match its brew counterpart for the form display.
func (GitHubCLI) Name() string { return "gh" }

// Description returns the picker blurb.
func (GitHubCLI) Description() string { return "GitHub from the terminal: PRs, issues, runs" }

// Install adds GitHub's apt source and key, then installs gh.
func (g GitHubCLI) Install(stdout, stderr io.Writer) error {
	const keyring = "/etc/apt/keyrings/githubcli-archive-keyring.gpg"
	steps := [][]string{
		SudoAptArgs([]string{"install", "-y", "ca-certificates", "curl"}),
		{"sh", "-c", "sudo -n mkdir -p -m 755 /etc/apt/keyrings && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo -n tee " + keyring + " > /dev/null && sudo -n chmod go+r " + keyring},
		{"sh", "-c", "echo \"deb [arch=$(dpkg --print-architecture) signed-by=" + keyring + "] https://cli.github.com/packages stable main\" | sudo -n tee /etc/apt/sources.list.d/github-cli.list > /dev/null"},
		SudoAptArgs([]string{"update"}),
		SudoAptArgs([]string{"install", "-y", "gh"}),
	}
	return runSteps("gh", steps, g.Run, stdout, stderr)
}

// Status reads the installed gh version via dpkg-query.
func (g GitHubCLI) Status() (pkg.InstallStatus, string, error) {
	return dpkgVersion(g.Run, "gh")
}

// runSteps drives each step through run (or the production runner when nil).
func runSteps(what string, steps [][]string, run CmdRunner, stdout, stderr io.Writer) error {
	if run == nil {
		run = DefaultCmdRunner()
	}
	for _, step := range steps {
		if err := run(step, stdout, stderr); err != nil {
			return fmt.Errorf("%s install step %q: %w", what, step, err)
		}
	}
	return nil
}
