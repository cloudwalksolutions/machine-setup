package profiles_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/profiles"
)

var _ = Describe("the active profile marker", func() {
	var dir string

	BeforeEach(func() {
		dir = filepath.Join(GinkgoT().TempDir(), "profiles")
	})

	It("is empty when nothing has been activated", func() {
		active, err := profiles.Active(dir)

		Expect(err).NotTo(HaveOccurred())
		Expect(active).To(BeEmpty())
	})

	It("reads back the alias SetActive wrote, creating the dir on first use", func() {
		Expect(profiles.SetActive(dir, "cws")).To(Succeed())

		active, err := profiles.Active(dir)

		Expect(err).NotTo(HaveOccurred())
		Expect(active).To(Equal("cws"))
	})
})
