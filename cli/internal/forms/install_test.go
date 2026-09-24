package forms_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/forms"
	"tars/internal/pkg"
)

var _ = Describe("ShowInstallForm headless", func() {
	BeforeEach(func() { GinkgoT().Setenv("TARS_NO_FORM", "1") })

	It("selects only the tools that are not installed yet", func() {
		selected, err := forms.ShowInstallForm([]pkg.ToolInfo{
			{Name: "neovim", Installed: true, Version: "0.12.0"},
			{Name: "byobu"},
			{Name: "fzf"},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(selected).To(Equal([]string{"byobu", "fzf"}))
	})
})
