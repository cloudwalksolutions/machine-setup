package registry_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
	"tars/internal/pkg/apt"
	"tars/internal/pkg/brew"
	"tars/internal/pkg/registry"
)

type fakeInstallable struct{ name string }

func (f fakeInstallable) Name() string                 { return f.name }
func (f fakeInstallable) Install(_, _ io.Writer) error { return nil }
func (f fakeInstallable) Status() (pkg.InstallStatus, string, error) {
	return pkg.StatusNotInstalled, "", nil
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

	It("appends caller-provided extras to every supported-OS registry", func() {
		extra := fakeInstallable{name: "my-extra"}
		factoryWithExtra := registry.NewRegistryFactory(
			brew.Runner(brewSpy.Run),
			spyKit(),
			extra,
		)

		Expect(factoryWithExtra.For("darwin").Names()).To(ContainElement("my-extra"))
		Expect(factoryWithExtra.For("linux").Names()).To(ContainElement("my-extra"))
	})

	It("does NOT include extras when the OS is unsupported", func() {
		extra := fakeInstallable{name: "my-extra"}
		factoryWithExtra := registry.NewRegistryFactory(nil, apt.Kit{}, extra)

		Expect(factoryWithExtra.For("plan9").Installables()).To(BeEmpty())
	})
})
