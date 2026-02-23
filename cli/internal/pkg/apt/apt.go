package apt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// aptNames maps brew-style package names to their apt equivalents.
// Packages not listed here use the same name on both managers.
var aptNames = map[string]string{
	"go":     "golang",
	"node":   "nodejs",
	"python": "python3",
}

// customInstallers maps package names to functions that install them
// outside of apt (e.g. AppImage, curl installer).
var customInstallers = map[string]func(stdout, stderr io.Writer) error{
	"neovim":   installNeovimAppImage,
	"starship": installStarship,
}

// Apt is an apt-backed package manager adapter for Debian-based systems.
type Apt struct {
	stdout io.Writer
	stderr io.Writer
}

// New returns an Apt adapter that streams output to stdout and stderr.
func New(stdout, stderr io.Writer) *Apt {
	return &Apt{stdout: stdout, stderr: stderr}
}

func (a *Apt) Install(name string) error {
	if installer, ok := customInstallers[name]; ok {
		return installer(a.stdout, a.stderr)
	}
	return a.run("install", "-y", resolve(name))
}

func (a *Apt) Uninstall(name string) error {
	return a.run("remove", "-y", resolve(name))
}

func (a *Apt) Update(name string) error {
	if _, ok := customInstallers[name]; ok {
		// For custom-installed packages, reinstall to update
		return a.Install(name)
	}
	return a.run("install", "-y", "--only-upgrade", resolve(name))
}

// IsInstalled returns true if the package is installed according to dpkg,
// or if a custom-installed binary is found in PATH.
func (a *Apt) IsInstalled(name string) (bool, error) {
	if _, ok := customInstallers[name]; ok {
		_, err := exec.LookPath(name)
		if err != nil {
			// neovim binary is "nvim", not "neovim"
			if name == "neovim" {
				_, err = exec.LookPath("nvim")
			}
		}
		return err == nil, nil
	}
	out, err := exec.Command("dpkg", "-l", resolve(name)).Output()
	if err != nil {
		return false, nil
	}
	return strings.Contains(string(out), "ii"), nil
}

// resolve translates a brew-style name to the apt package name.
func resolve(name string) string {
	if mapped, ok := aptNames[name]; ok {
		return mapped
	}
	return name
}

func (a *Apt) run(args ...string) error {
	cmd := exec.Command("sudo", append([]string{"apt"}, args...)...)
	cmd.Stdout = a.stdout
	cmd.Stderr = a.stderr
	return cmd.Run()
}

// installNeovimAppImage downloads the nvim AppImage to ~/.local/bin/nvim.
func installNeovimAppImage(stdout, stderr io.Writer) error {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	} else if arch == "arm64" {
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

	if err := os.Chmod(dest, 0o755); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	fmt.Fprintf(stdout, "Neovim AppImage installed to %s\n", dest)
	return nil
}

// installStarship downloads starship via the official installer script.
func installStarship(stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, "Installing starship via official installer...")
	cmd := exec.Command("sh", "-c", "curl -sS https://starship.rs/install.sh | sh -s -- -y")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
