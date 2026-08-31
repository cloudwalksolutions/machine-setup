package apt_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/pkg/apt"
)

var _ = Describe("SudoAptArgs", func() {
	It("builds a non-interactive sudo apt-get argv", func() {
		Expect(apt.SudoAptArgs([]string{"install", "-y", "jq"})).To(Equal([]string{
			"sudo", "-n", "DEBIAN_FRONTEND=noninteractive", "apt-get", "install", "-y", "jq",
		}))
	})
})

var _ = Describe("Package.Install", func() {
	It("resolves brew-style names to their apt equivalents", func() {
		var gotArgs []string
		spy := func(args []string, _, _ io.Writer) error {
			gotArgs = args
			return nil
		}

		Expect(apt.NewPackage("go", spy).Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())
		Expect(gotArgs).To(Equal([]string{"install", "-y", "golang"}))

		Expect(apt.NewPackage("node", spy).Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())
		Expect(gotArgs).To(Equal([]string{"install", "-y", "nodejs"}))

		Expect(apt.NewPackage("python", spy).Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())
		Expect(gotArgs).To(Equal([]string{"install", "-y", "python3"}))
	})

	It("invokes the runner with [install -y <name>] for a name without mapping", func() {
		var gotArgs []string
		spy := func(args []string, _, _ io.Writer) error {
			gotArgs = args
			return nil
		}

		err := apt.NewPackage("byobu", spy).Install(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(gotArgs).To(Equal([]string{"install", "-y", "byobu"}))
	})
})

// fakeNvimTarball builds an in-memory tar.gz shaped like the upstream release:
// one top-level dir containing bin/nvim (0755) and a share file.
func fakeNvimTarball(topDir string) io.ReadCloser {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	write := func(name string, mode int64, body string) {
		hdr := &tar.Header{Name: name, Mode: mode, Size: int64(len(body))}
		ExpectWithOffset(1, tw.WriteHeader(hdr)).To(Succeed())
		_, err := tw.Write([]byte(body))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}
	ExpectWithOffset(1, tw.WriteHeader(&tar.Header{
		Name: topDir + "/", Typeflag: tar.TypeDir, Mode: 0o755,
	})).To(Succeed())
	write(topDir+"/bin/nvim", 0o755, "ELF-FAKE")
	write(topDir+"/share/nvim/runtime/doc.txt", 0o644, "DOCS")

	ExpectWithOffset(1, tw.Close()).To(Succeed())
	ExpectWithOffset(1, gz.Close()).To(Succeed())
	return io.NopCloser(&buf)
}

var _ = Describe("NeovimTarball", func() {
	It("downloads the arch tarball, extracts it to ~/.local/nvim, and links ~/.local/bin/nvim", func() {
		home := GinkgoT().TempDir()
		var gotURL string
		fetch := func(url string) (io.ReadCloser, error) {
			gotURL = url
			return fakeNvimTarball("nvim-linux-arm64"), nil
		}

		nv := apt.NeovimTarball{Fetch: fetch, Home: home, Arch: "arm64"}
		Expect(nv.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())

		Expect(gotURL).To(ContainSubstring("nvim-linux-arm64.tar.gz"))

		bin := filepath.Join(home, ".local", "nvim", "bin", "nvim")
		data, err := os.ReadFile(bin)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("ELF-FAKE"))
		info, err := os.Stat(bin)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()&0o111).NotTo(BeZero(), "extracted nvim must keep its exec bit")

		link := filepath.Join(home, ".local", "bin", "nvim")
		target, err := os.Readlink(link)
		Expect(err).NotTo(HaveOccurred())
		Expect(target).To(Equal(bin))
	})

	It("requests the x86_64 asset for amd64", func() {
		var gotURL string
		fetch := func(url string) (io.ReadCloser, error) {
			gotURL = url
			return fakeNvimTarball("nvim-linux-x86_64"), nil
		}

		nv := apt.NeovimTarball{Fetch: fetch, Home: GinkgoT().TempDir(), Arch: "amd64"}
		Expect(nv.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())

		Expect(gotURL).To(ContainSubstring("nvim-linux-x86_64.tar.gz"))
	})

	It("propagates a fetch failure", func() {
		fetch := func(string) (io.ReadCloser, error) { return nil, errors.New("HTTP 404") }
		nv := apt.NeovimTarball{Fetch: fetch, Home: GinkgoT().TempDir(), Arch: "amd64"}

		Expect(nv.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(MatchError(ContainSubstring("404")))
	})
})

var _ = Describe("GitHubCLI", func() {
	It("reports the name gh", func() {
		Expect(apt.GitHubCLI{}.Name()).To(Equal("gh"))
	})

	It("adds GitHub's apt repo non-interactively, then installs gh", func() {
		var steps [][]string
		spy := func(argv []string, _, _ io.Writer) error {
			steps = append(steps, argv)
			return nil
		}

		Expect(apt.NewGitHubCLI(spy).Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())

		Expect(steps).NotTo(BeEmpty())
		joined := ""
		for _, s := range steps {
			joined += " " + fmt.Sprint(s)
		}
		Expect(joined).To(ContainSubstring("cli.github.com/packages"))
		Expect(joined).To(ContainSubstring("arch=$(dpkg --print-architecture)"))
		Expect(joined).To(ContainSubstring("DEBIAN_FRONTEND=noninteractive"))
		last := steps[len(steps)-1]
		Expect(last[len(last)-1]).To(Equal("gh"))
	})
})

var _ = Describe("GCloudCLI", func() {
	It("reports the name gcloud", func() {
		Expect(apt.GCloudCLI{}.Name()).To(Equal("gcloud"))
	})

	It("drives the injected runner through the documented apt steps", func() {
		var steps [][]string
		spy := func(argv []string, _, _ io.Writer) error {
			steps = append(steps, argv)
			return nil
		}

		err := apt.NewGCloudCLI(spy).Install(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(steps).To(HaveLen(6))
		Expect(steps[0]).To(Equal(apt.SudoAptArgs([]string{"update"})))
		Expect(steps[5]).To(Equal(apt.SudoAptArgs([]string{"install", "-y", "google-cloud-cli"})))
		joined := fmt.Sprint(steps)
		Expect(joined).To(ContainSubstring("arch=$(dpkg --print-architecture)"))
	})
})
