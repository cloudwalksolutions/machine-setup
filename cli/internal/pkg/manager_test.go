package pkg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/pkg"
)

var _ = Describe("InstallStatus", func() {
	It("reports correct string representation", func() {
		Expect(pkg.StatusNotInstalled.String()).To(Equal("not installed"))
		Expect(pkg.StatusUpdateAvailable.String()).To(Equal("update available"))
		Expect(pkg.StatusUpToDate.String()).To(Equal("up to date"))
		Expect(pkg.InstallStatus(99).String()).To(Equal("unknown"))
	})
})
