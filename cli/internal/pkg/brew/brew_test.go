package brew_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/pkg/brew"
)

func skipUnlessIntegration() {
	if os.Getenv("INTEGRATION") == "" {
		Skip("set INTEGRATION=1 to run brew integration tests")
	}
}

var _ = Describe("Formula.Install (integration)", Ordered, func() {
	BeforeAll(func() {
		skipUnlessIntegration()
		// Ensure clean state.
		_ = exec.Command("brew", "uninstall", "hello").Run()
	})

	AfterAll(func() {
		if os.Getenv("INTEGRATION") == "" {
			return
		}
		_ = exec.Command("brew", "uninstall", "hello").Run()
	})

	It("installs hello via real brew and the binary becomes executable", func() {
		skipUnlessIntegration()

		err := brew.NewFormula("hello", brew.DefaultRunner()).
			Install(GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		path, err := exec.LookPath("hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(path).NotTo(BeEmpty())

		out, err := exec.Command("hello").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(ContainSubstring("Hello, world!"))
	})
})
