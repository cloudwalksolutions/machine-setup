package registry_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
	"tars/internal/pkg/apt"
	"tars/internal/pkg/brew"
	"tars/internal/pkg/registry"
)

type fakeInstallable struct {
	name        string
	description string
	version     string
}

func (f fakeInstallable) Name() string                 { return f.name }
func (f fakeInstallable) Description() string          { return f.description }
func (f fakeInstallable) Install(_, _ io.Writer) error { return nil }
func (f fakeInstallable) Status() (pkg.InstallStatus, string, error) {
	if f.version != "" {
		return pkg.StatusUpToDate, f.version, nil
	}
	return pkg.StatusNotInstalled, "", nil
}

type slowInstallable struct {
	name  string
	delay time.Duration
}

func (s slowInstallable) Name() string                 { return s.name }
func (s slowInstallable) Description() string          { return "" }
func (s slowInstallable) Install(_, _ io.Writer) error { return nil }
func (s slowInstallable) Status() (pkg.InstallStatus, string, error) {
	time.Sleep(s.delay)
	return pkg.StatusUpToDate, "1.0", nil
}

var _ = Describe("DevToolRegistry", func() {
	var reg *registry.DevToolRegistry

	BeforeEach(func() {
		reg = registry.NewDevToolRegistry()
	})

	It("is empty on construction", func() {
		Expect(reg.Installables()).To(BeEmpty())
	})

	It("Add appends an installable, preserving declaration order", func() {
		reg.Add(fakeInstallable{name: "a"})
		reg.Add(fakeInstallable{name: "b"})

		Expect(reg.Installables()).To(HaveLen(2))
		Expect(reg.Installables()[0].Name()).To(Equal("a"))
		Expect(reg.Installables()[1].Name()).To(Equal("b"))
	})

	It("AddAll appends a batch in order", func() {
		reg.AddAll([]pkg.Installable{
			fakeInstallable{name: "x"},
			fakeInstallable{name: "y"},
			fakeInstallable{name: "z"},
		})

		Expect(reg.Installables()).To(HaveLen(3))
		Expect(reg.Installables()[2].Name()).To(Equal("z"))
	})

	It("Catalog projects each installable's name, description, installed state and version in order", func() {
		reg := registry.NewDevToolRegistry().
			Add(fakeInstallable{name: "neovim", description: "the editor", version: "0.12.5"}).
			Add(fakeInstallable{name: "fzf", description: "the finder"})

		Expect(reg.Catalog()).To(Equal([]pkg.ToolInfo{
			{Name: "neovim", Description: "the editor", Installed: true, Version: "0.12.5"},
			{Name: "fzf", Description: "the finder"},
		}))
	})

	It("Catalog queries statuses concurrently so slow package managers do not stall the form", func() {
		for _, name := range []string{"a", "b", "c", "d", "e"} {
			reg.Add(slowInstallable{name: name, delay: 100 * time.Millisecond})
		}

		start := time.Now()
		infos := reg.Catalog()

		Expect(time.Since(start)).To(BeNumerically("<", 300*time.Millisecond))
		Expect(infos).To(HaveLen(5))
		Expect(infos[4].Name).To(Equal("e"))
	})

	It("Names projects each installable's name in order", func() {
		reg.Add(fakeInstallable{name: "foo"}).Add(fakeInstallable{name: "bar"})

		Expect(reg.Names()).To(Equal([]string{"foo", "bar"}))
	})
})

type recordingRunner struct {
	lastArgs []string
	calls    int
}

func (r *recordingRunner) Run(args []string, _, _ io.Writer) error {
	r.lastArgs = args
	r.calls++
	return nil
}

// nothingOnPath keeps specs hermetic: no binary is ever found on PATH.
var nothingOnPath = pkg.PathProbe{LookPath: func(string) (string, error) { return "", errors.New("not found") }}

func byName(reg *registry.DevToolRegistry, name string) pkg.Installable {
	for _, t := range reg.Installables() {
		if t.Name() == name {
			return t
		}
	}
	Fail("no installable named " + name)
	return nil
}

var _ = Describe("RegistryFactory", func() {
	var (
		brewSpy    *recordingRunner
		aptSpy     *recordingRunner
		cmdSpy     *recordingRunner
		fetchCalls int
		factory    registry.RegistryFactory
	)

	minimalTarGz := func() io.ReadCloser {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gz)
		Expect(tw.Close()).To(Succeed())
		Expect(gz.Close()).To(Succeed())
		return io.NopCloser(&buf)
	}

	spyKit := func() apt.Kit {
		return apt.Kit{
			Apt: aptSpy.Run,
			Cmd: cmdSpy.Run,
			Fetch: func(string) (io.ReadCloser, error) {
				fetchCalls++
				return minimalTarGz(), nil
			},
			Home: GinkgoT().TempDir(),
		}
	}

	BeforeEach(func() {
		brewSpy = &recordingRunner{}
		aptSpy = &recordingRunner{}
		cmdSpy = &recordingRunner{}
		fetchCalls = 0
		factory = registry.NewRegistryFactory(
			brew.Runner(brewSpy.Run),
			spyKit(),
			nothingOnPath,
		)
	})

	It("returns an empty registry for an unsupported OS", func() {
		Expect(factory.For("plan9").Installables()).To(BeEmpty())
	})

	It("on darwin, populates installables whose installs route through brew", func() {
		reg := factory.For("darwin")
		Expect(reg.Installables()).NotTo(BeEmpty())

		Expect(reg.Installables()[0].Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())

		Expect(brewSpy.calls).To(BeNumerically(">", 0))
		Expect(aptSpy.calls).To(Equal(0))
	})

	It("on darwin, asks PATH for each tool's binary when brew does not list it", func() {
		brewMissing := func(_ []string, _, _ io.Writer) error { return errors.New("exit status 1") }
		var asked []string
		probe := pkg.PathProbe{LookPath: func(file string) (string, error) {
			asked = append(asked, file)
			return "", errors.New("not found")
		}}
		reg := registry.NewRegistryFactory(brew.Runner(brewMissing), spyKit(), probe).For("darwin")

		for _, name := range []string{"gcloud-cli", "neovim", "ripgrep", "python", "terraform", "byobu"} {
			_, _, err := byName(reg, name).Status()
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(asked).To(Equal([]string{"gcloud", "nvim", "rg", "python3", "terraform", "byobu"}))
	})

	It("on darwin, lists the formulas by popularity: editors and terminal first, fzf/ripgrep last, then languages, then devops", func() {
		names := factory.For("darwin").Names()

		Expect(names[:23]).To(Equal([]string{
			"neovim", "tree-sitter-cli", "byobu", "gh", "lazygit", "jq", "bat", "eza", "k9s", "lazydocker", "k3d", "golangci-lint", "fzf", "ripgrep",
			"go", "node", "python", "ruby", "rustup", "ghcup", "yarn", "n",
			"ansible",
		}))
	})

	It("on darwin, rustup bootstraps the stable toolchain right after brew installs it", func() {
		var steps [][]string
		factory = registry.NewRegistryFactory(brew.Runner(brewSpy.Run), apt.Kit{
			Cmd: func(argv []string, _, _ io.Writer) error { steps = append(steps, argv); return nil },
		}, nothingOnPath)

		tool := byName(factory.For("darwin"), "rustup")
		Expect(tool.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())

		Expect(brewSpy.lastArgs).To(Equal([]string{"install", "rustup"}))
		Expect(steps).To(Equal([][]string{
			{"rustup", "install", "stable"},
			{"rustup", "default", "stable"},
		}))
	})

	It("on darwin, ghcup installs and selects the recommended GHC right after brew installs it", func() {
		var steps [][]string
		factory = registry.NewRegistryFactory(brew.Runner(brewSpy.Run), apt.Kit{
			Cmd: func(argv []string, _, _ io.Writer) error { steps = append(steps, argv); return nil },
		}, nothingOnPath)

		tool := byName(factory.For("darwin"), "ghcup")
		Expect(tool.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())

		Expect(brewSpy.lastArgs).To(Equal([]string{"install", "ghcup"}))
		Expect(steps).To(Equal([][]string{
			{"ghcup", "install", "ghc", "recommended"},
			{"ghcup", "set", "ghc", "recommended"},
		}))
	})

	It("on linux, lists tools in the same popularity order", func() {
		Expect(factory.For("linux").Names()).To(Equal([]string{
			"neovim", "tree-sitter-cli", "byobu", "gh", "jq", "bat", "fzf", "ripgrep",
			"go", "node", "python",
			"gcloud",
		}))
	})

	It("gives every tool on both platforms a concise description", func() {
		var tools []pkg.Installable
		tools = append(tools, factory.For("darwin").Installables()...)
		tools = append(tools, factory.For("linux").Installables()...)

		for _, t := range tools {
			Expect(t.Description()).NotTo(BeEmpty(), t.Name())
			Expect(len(t.Description())).To(BeNumerically("<=", 48), t.Name())
		}
	})

	It("on darwin, includes the gcloud-cli cask", func() {
		Expect(factory.For("darwin").Names()).To(ContainElement("gcloud-cli"))
	})

	It("on linux, includes a gcloud installable", func() {
		Expect(factory.For("linux").Names()).To(ContainElement("gcloud"))
	})

	It("on linux, at least one installable routes through apt", func() {
		reg := factory.For("linux")
		Expect(reg.Installables()).NotTo(BeEmpty())

		for _, tool := range reg.Installables() {
			aptSpy.lastArgs = nil
			_ = tool.Install(&bytes.Buffer{}, &bytes.Buffer{})
			if len(aptSpy.lastArgs) > 0 {
				return
			}
		}
		Fail("no installable in the linux registry routed through apt")
	})

	It("on linux, gh installs via the GitHub apt-repo steps, not a plain apt install", func() {
		reg := factory.For("linux")

		for _, tool := range reg.Installables() {
			if tool.Name() != "gh" {
				continue
			}
			Expect(tool.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())
			Expect(cmdSpy.calls).To(BeNumerically(">", 0), "gh must run repo-setup steps")
			Expect(aptSpy.calls).To(Equal(0), "gh must not be a plain apt package")
			return
		}
		Fail("no gh installable in the linux registry")
	})

	It("on linux, every installable routes through an injected seam (airgapped unit contract)", func() {
		reg := factory.For("linux")
		Expect(reg.Installables()).NotTo(BeEmpty())

		for _, tool := range reg.Installables() {
			aptBefore, cmdBefore, fetchBefore := aptSpy.calls, cmdSpy.calls, fetchCalls

			Expect(tool.Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed(), tool.Name())

			touched := aptSpy.calls > aptBefore || cmdSpy.calls > cmdBefore || fetchCalls > fetchBefore
			Expect(touched).To(BeTrue(), "%s performed no seam call — likely real I/O", tool.Name())
		}
	})

	It("on linux, every installable reports its status through the injected command seam", func() {
		kit := spyKit()
		Expect(os.MkdirAll(filepath.Join(kit.Home, ".local", "nvim"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(kit.Home, ".local", "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(kit.Home, ".local", "bin", "tree-sitter"), nil, 0o755)).To(Succeed())
		reg := registry.NewRegistryFactory(brew.Runner(brewSpy.Run), kit, nothingOnPath).For("linux")

		for _, tool := range reg.Installables() {
			before := cmdSpy.calls
			_, _, err := tool.Status()
			Expect(err).NotTo(HaveOccurred(), tool.Name())
			Expect(cmdSpy.calls).To(BeNumerically(">", before), "%s did not query through the seam", tool.Name())
		}
	})

	It("appends caller-provided extras to every supported-OS registry", func() {
		extra := fakeInstallable{name: "my-extra"}
		factoryWithExtra := registry.NewRegistryFactory(
			brew.Runner(brewSpy.Run),
			spyKit(),
			nothingOnPath,
			extra,
		)

		Expect(factoryWithExtra.For("darwin").Names()).To(ContainElement("my-extra"))
		Expect(factoryWithExtra.For("linux").Names()).To(ContainElement("my-extra"))
	})

	It("does NOT include extras when the OS is unsupported", func() {
		extra := fakeInstallable{name: "my-extra"}
		factoryWithExtra := registry.NewRegistryFactory(nil, apt.Kit{}, nothingOnPath, extra)

		Expect(factoryWithExtra.For("plan9").Installables()).To(BeEmpty())
	})
})
