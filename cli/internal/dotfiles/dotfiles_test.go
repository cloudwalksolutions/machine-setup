package dotfiles_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func repoRoot() string {
	abs, err := filepath.Abs(filepath.Join("..", "..", ".."))
	Expect(err).NotTo(HaveOccurred())
	return abs
}

func zshFiles() []string {
	entries, err := os.ReadDir(filepath.Join(repoRoot(), "zsh"))
	Expect(err).NotTo(HaveOccurred())
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, filepath.Join(repoRoot(), "zsh", e.Name()))
		}
	}
	Expect(out).NotTo(BeEmpty())
	return out
}

func deployedZshFiles() []string {
	var out []string
	for _, f := range zshFiles() {
		if !strings.HasSuffix(f, ".template") {
			out = append(out, f)
		}
	}
	return out
}

func read(path string) string {
	b, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}

var _ = Describe("shipped terminal font", func() {
	fontSpec := func() string {
		return strings.TrimSpace(read(filepath.Join(repoRoot(), "terminal", "font")))
	}

	It("is a PostScript name and a size, so iTerm2 can resolve it", func() {
		Expect(fontSpec()).To(MatchRegexp(`^\S+ \d+$`))
	})

	It("names a face that the shipped font files actually provide", func() {
		fcScan, err := exec.LookPath("fc-scan")
		if err != nil {
			Skip("fontconfig not installed")
		}
		name, _, _ := strings.Cut(fontSpec(), " ")

		out, err := exec.Command(fcScan, "--format", "%{postscriptname}\n",
			filepath.Join(repoRoot(), "fonts")).CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		Expect(strings.Split(string(out), "\n")).To(ContainElement(name),
			"terminal/font names %q but fonts/ provides:\n%s", name, out)
	})
})

var _ = Describe("shipped zsh dotfiles", func() {
	It("export EDITOR, which the vim aliases interpolate at source time", func() {
		Expect(read(filepath.Join(repoRoot(), "zsh", "profile"))).
			To(MatchRegexp(`(?m)^\s*export EDITOR=`))
	})

	It("define no alias with backticks, which would run on every shell start", func() {
		backtickAlias := regexp.MustCompile("(?m)^\\s*alias\\s+[^=]+=`")
		for _, f := range zshFiles() {
			Expect(backtickAlias.MatchString(read(f))).To(BeFalse(), f)
		}
	})

	It("carry no email address outside the placeholder templates", func() {
		email := regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.]+`)
		for _, f := range deployedZshFiles() {
			Expect(email.FindString(read(f))).To(BeEmpty(),
				"%s contains an email address; move it to ~/.zshrc_secret", f)
		}
	})

	It("hardcode no home directory", func() {
		for _, f := range zshFiles() {
			Expect(read(f)).NotTo(ContainSubstring("/Users/"), f)
		}
	})

	It("source cleanly when no machine-local override is present", func() {
		zsh, err := exec.LookPath("zsh")
		if err != nil {
			Skip("zsh not installed")
		}
		cmd := exec.Command(zsh, "-f", "-c",
			"source "+filepath.Join(repoRoot(), "zsh", "profile"))
		cmd.Env = append(os.Environ(), "HOME="+GinkgoT().TempDir())

		out, err := cmd.CombinedOutput()

		Expect(err).NotTo(HaveOccurred(), "%s", out)
	})

	It("are syntactically valid zsh", func() {
		zsh, err := exec.LookPath("zsh")
		if err != nil {
			Skip("zsh not installed")
		}
		for _, f := range deployedZshFiles() {
			out, err := exec.Command(zsh, "-n", f).CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "%s: %s", f, out)
		}
	})
})
