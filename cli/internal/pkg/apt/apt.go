// Package apt provides polymorphic Installable implementations for the
// Debian/Ubuntu side of the CLI's curated install list. Each installable
// type (Package, NeovimAppImage) is its own kind that knows how to install
// itself — no dispatcher map, no type switches.
package apt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Runner runs an apt subcommand. Production wiring shells out to
// `sudo apt …`; tests inject a recorder.
type Runner func(args []string, stdout, stderr io.Writer) error

// DefaultRunner returns the production Runner.
func DefaultRunner() Runner {
	return func(args []string, stdout, stderr io.Writer) error {
		cmd := exec.Command("sudo", append([]string{"apt"}, args...)...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
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
	name string
	run  Runner
}

// NewPackage returns a Package bound to a runner.
func NewPackage(name string, run Runner) Package {
	return Package{name: name, run: run}
}

// Name returns the brew-style name (unresolved). This is what the user sees.
func (p Package) Name() string { return p.name }

// Install runs `apt install -y <resolved-name>`.
func (p Package) Install(stdout, stderr io.Writer) error {
	resolved := p.name
	if mapped, ok := aptNames[p.name]; ok {
		resolved = mapped
	}
	return p.run([]string{"install", "-y", resolved}, stdout, stderr)
}

// NeovimAppImage installs Neovim by downloading the upstream AppImage to
// ~/.local/bin/nvim. Used on Linux where the apt package is often outdated.
type NeovimAppImage struct{}

// Name reports "neovim" to match its brew counterpart for the form display.
func (NeovimAppImage) Name() string { return "neovim" }

// Install downloads the AppImage and makes it executable.
func (NeovimAppImage) Install(stdout, stderr io.Writer) error {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	}

	url := fmt.Sprintf("https://github.com/neovim/neovim/releases/download/v0.11.6/nvim-linux-%s.appimage", arch)
	dest := filepath.Join(os.Getenv("HOME"), ".local", "bin", "nvim")

	fmt.Fprintf(stdout, "Downloading Neovim AppImage to %s...\n", dest)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("downloading neovim: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return os.Chmod(dest, 0o755)
}

// GCloudCLI installs the Google Cloud CLI on Debian/Ubuntu by adding Google's
// apt repository (the package is not in the default archives) and installing
// google-cloud-cli. Mirrors the documented steps at
// https://cloud.google.com/sdk/docs/install#deb.
type GCloudCLI struct{}

// Name reports "gcloud" to match its brew-cask counterpart for the form display.
func (GCloudCLI) Name() string { return "gcloud" }

// Install adds Google's apt source and key, then installs google-cloud-cli.
func (GCloudCLI) Install(stdout, stderr io.Writer) error {
	const keyring = "/usr/share/keyrings/cloud.google.gpg"
	steps := [][]string{
		{"sudo", "apt-get", "update"},
		{"sudo", "apt-get", "install", "-y", "apt-transport-https", "ca-certificates", "gnupg", "curl"},
		// Fetch Google's signing key and dearmor it into the keyring.
		{"sh", "-c", "curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o " + keyring},
		// Register the cloud-sdk apt source, pinned to the keyring.
		{"sh", "-c", "echo 'deb [signed-by=" + keyring + "] https://packages.cloud.google.com/apt cloud-sdk main' | sudo tee /etc/apt/sources.list.d/google-cloud-sdk.list"},
		{"sudo", "apt-get", "update"},
		{"sudo", "apt-get", "install", "-y", "google-cloud-cli"},
	}
	for _, step := range steps {
		cmd := exec.Command(step[0], step[1:]...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("gcloud install step %q: %w", step, err)
		}
	}
	return nil
}

