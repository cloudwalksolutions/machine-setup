package apt

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func skipUnlessIntegration() {
	if os.Getenv("INTEGRATION") == "" {
		Skip("set INTEGRATION=1 to run apt integration tests")
	}
}

var _ = Describe("Apt manager (integration)", Ordered, func() {
	var a *Apt

	BeforeAll(func() {
		skipUnlessIntegration()
		a = New(GinkgoWriter, GinkgoWriter)
		_ = a.Uninstall("hello")
	})

	It("reports hello is not installed before test", func() {
		skipUnlessIntegration()
		installed, err := a.IsInstalled("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeFalse())
	})

	It("installs hello and the binary is executable", func() {
		skipUnlessIntegration()
		Expect(a.Install("hello")).To(Succeed())

		path, err := exec.LookPath("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).NotTo(BeEmpty())

		out, err := exec.Command("hello").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(ContainSubstring("Hello, world!"))
	})

	It("reports hello is installed after install", func() {
		skipUnlessIntegration()
		installed, err := a.IsInstalled("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())
	})

	It("upgrades hello without error", func() {
		skipUnlessIntegration()
		Expect(a.Update("hello")).To(Succeed())
	})

	It("uninstalls hello", func() {
		skipUnlessIntegration()
		Expect(a.Uninstall("hello")).To(Succeed())
		_, err := exec.LookPath("hello")
		Expect(err).To(HaveOccurred())
	})

	It("reports hello is not installed after uninstall", func() {
		skipUnlessIntegration()
		installed, err := a.IsInstalled("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeFalse())
	})
})

var _ = Describe("resolve", func() {
	DescribeTable("maps brew names to apt names",
		func(input, expected string) {
			Expect(resolve(input)).To(Equal(expected))
		},
		Entry("go → golang", "go", "golang"),
		Entry("node → nodejs", "node", "nodejs"),
		Entry("python → python3", "python", "python3"),
		Entry("neovim unchanged", "neovim", "neovim"),
		Entry("byobu unchanged", "byobu", "byobu"),
		Entry("fzf unchanged", "fzf", "fzf"),
	)
})

var _ = Describe("customInstallers", func() {
	DescribeTable("routes packages to custom installers",
		func(name string, expectCustom bool) {
			_, ok := customInstallers[name]
			Expect(ok).To(Equal(expectCustom))
		},
		Entry("neovim has custom installer", "neovim", true),
		Entry("starship has custom installer", "starship", true),
		Entry("byobu uses apt", "byobu", false),
		Entry("fzf uses apt", "fzf", false),
	)
})
