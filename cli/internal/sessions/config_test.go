package sessions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/sessions"
)

var _ = Describe("WindowName", func() {
	It("uses the dir basename with '.' and ':' sanitized to '_'", func() {
		Expect(sessions.WindowName("/projects/machine-setup")).To(Equal("machine-setup"))
		Expect(sessions.WindowName("/projects/api.v2:beta")).To(Equal("api_v2_beta"))
	})
})
