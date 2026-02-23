package brew

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func skipUnlessIntegration() {
	if os.Getenv("INTEGRATION") == "" {
		Skip("set INTEGRATION=1 to run brew integration tests")
	}
}

var _ = Describe("Brew manager", Ordered, func() {
	var b *Brew

	BeforeAll(func() {
		skipUnlessIntegration()
		b = New(GinkgoWriter, GinkgoWriter)
		// ensure clean state — ignore error if not installed
		_ = b.Uninstall("hello")
	})

	It("reports hello is not installed before test", func() {
		skipUnlessIntegration()
		installed, err := b.IsInstalled("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeFalse())

		// binary must not be resolvable in PATH
		_, lookErr := exec.LookPath("hello")
		Expect(lookErr).To(HaveOccurred())
	})

	It("installs hello and the binary is executable", func() {
		skipUnlessIntegration()
		Expect(b.Install("hello")).To(Succeed())

		// binary must be resolvable in PATH
		path, err := exec.LookPath("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).NotTo(BeEmpty())

		// running the binary must produce the expected greeting
		out, err := exec.Command("hello").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(ContainSubstring("Hello, world!"))
	})

	It("reports hello is installed after install", func() {
		skipUnlessIntegration()
		installed, err := b.IsInstalled("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeTrue())
	})

	It("upgrades hello without error", func() {
		skipUnlessIntegration()
		Expect(b.Update("hello")).To(Succeed())

		// binary must still be executable after upgrade
		out, err := exec.Command("hello").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(ContainSubstring("Hello, world!"))
	})

	It("uninstalls hello", func() {
		skipUnlessIntegration()
		Expect(b.Uninstall("hello")).To(Succeed())

		// binary must be gone from PATH
		_, lookErr := exec.LookPath("hello")
		Expect(lookErr).To(HaveOccurred())
	})

	It("reports hello is not installed after uninstall", func() {
		skipUnlessIntegration()
		installed, err := b.IsInstalled("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(installed).To(BeFalse())
	})
})
