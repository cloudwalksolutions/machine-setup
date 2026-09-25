package forms_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/forms"
	"tars/internal/profiles"
)

var _ = Describe("standalone pickers with TARS_NO_FORM", func() {
	BeforeEach(func() { GinkgoT().Setenv("TARS_NO_FORM", "1") })

	It("session picker takes the first option and rejects an empty list", func() {
		Expect(forms.ShowSessionPicker([]string{"work", "(all)"})).To(Equal("work"))
		_, err := forms.ShowSessionPicker(nil)
		Expect(err).To(HaveOccurred())
	})

	It("profile form returns the defaults unchanged", func() {
		p := profiles.Profile{Name: "acme", Alias: "ac", Email: "me@acme.io"}
		Expect(forms.ShowProfileForm(p)).To(Equal(p))
	})
})
