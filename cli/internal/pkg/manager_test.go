package pkg_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/pkg"
	"github.com/cloudwalk/machine-setup/internal/pkg/apt"
	"github.com/cloudwalk/machine-setup/internal/pkg/brew"
)

// fakeInstallable is a minimal pkg.Installable for exercising DevToolRegistry.
// Tests should NOT use it to verify behavior of real Installable types — those
// are tested in their own packages.
type fakeInstallable struct{ name string }

func (f fakeInstallable) Name() string                    { return f.name }
func (f fakeInstallable) Install(_, _ io.Writer) error    { return nil }

var _ = Describe("DevToolRegistry", func() {
	var registry *pkg.DevToolRegistry

	BeforeEach(func() {
		registry = pkg.NewDevToolRegistry()
	})

	It("is empty on construction", func() {
		Expect(registry.Installables()).To(BeEmpty())
	})

	It("Add appends an installable, preserving declaration order", func() {
		registry.Add(fakeInstallable{name: "a"})
		registry.Add(fakeInstallable{name: "b"})

		Expect(registry.Installables()).To(HaveLen(2))
		Expect(registry.Installables()[0].Name()).To(Equal("a"))
		Expect(registry.Installables()[1].Name()).To(Equal("b"))
	})

	It("AddAll appends a batch in order", func() {
		registry.AddAll([]pkg.Installable{
			fakeInstallable{name: "x"},
			fakeInstallable{name: "y"},
			fakeInstallable{name: "z"},
		})

		Expect(registry.Installables()).To(HaveLen(3))
		Expect(registry.Installables()[2].Name()).To(Equal("z"))
	})

	It("Names projects each installable's name in order", func() {
		registry.Add(fakeInstallable{name: "foo"}).Add(fakeInstallable{name: "bar"})

		Expect(registry.Names()).To(Equal([]string{"foo", "bar"}))
	})
})

// recordingRunner records the args of every brew/apt invocation that flows
// through a registry-emitted Installable.
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
		brewSpy *recordingRunner
		aptSpy  *recordingRunner
		factory pkg.RegistryFactory
	)

	BeforeEach(func() {
		brewSpy = &recordingRunner{}
		aptSpy = &recordingRunner{}
		factory = pkg.NewRegistryFactory(
			brew.Runner(brewSpy.Run),
			apt.Runner(aptSpy.Run),
		)
	})

	It("returns an empty registry for an unsupported OS", func() {
		Expect(factory.For("plan9").Installables()).To(BeEmpty())
	})

	It("on darwin, populates installables whose installs route through brew", func() {
		registry := factory.For("darwin")
		Expect(registry.Installables()).NotTo(BeEmpty())

		Expect(registry.Installables()[0].Install(&bytes.Buffer{}, &bytes.Buffer{})).To(Succeed())

		Expect(brewSpy.calls).To(BeNumerically(">", 0))
		Expect(aptSpy.calls).To(Equal(0))
	})

	It("on linux, at least one installable routes through apt", func() {
		registry := factory.For("linux")
		Expect(registry.Installables()).NotTo(BeEmpty())

		for _, tool := range registry.Installables() {
			aptSpy.lastArgs = nil
			_ = tool.Install(&bytes.Buffer{}, &bytes.Buffer{})
			if len(aptSpy.lastArgs) > 0 {
				return
			}
		}
		Fail("no installable in the linux registry routed through apt")
	})

	It("appends caller-provided extras to every supported-OS registry", func() {
		extra := fakeInstallable{name: "my-extra"}
		factoryWithExtra := pkg.NewRegistryFactory(
			brew.Runner(brewSpy.Run),
			apt.Runner(aptSpy.Run),
			extra,
		)

		Expect(factoryWithExtra.For("darwin").Names()).To(ContainElement("my-extra"))
		Expect(factoryWithExtra.For("linux").Names()).To(ContainElement("my-extra"))
	})

	It("does NOT include extras when the OS is unsupported", func() {
		extra := fakeInstallable{name: "my-extra"}
		factoryWithExtra := pkg.NewRegistryFactory(nil, nil, extra)

		Expect(factoryWithExtra.For("plan9").Installables()).To(BeEmpty())
	})
})
