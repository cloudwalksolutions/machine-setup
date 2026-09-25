package report_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/report"
)

var _ = Describe("LineWriter", func() {
	It("emits complete lines only, treats carriage returns as line breaks, and flushes the rest on Close", func() {
		var lines []string
		w := &report.LineWriter{Emit: func(s string) { lines = append(lines, s) }}

		_, _ = w.Write([]byte("Downloading 10%\rDownloading 50%\nInst"))
		_, _ = w.Write([]byte("alled\n\ntrailing"))
		Expect(lines).To(Equal([]string{"Downloading 10%", "Downloading 50%", "Installed"}))

		Expect(w.Close()).To(Succeed())
		Expect(lines).To(Equal([]string{"Downloading 10%", "Downloading 50%", "Installed", "trailing"}))
	})
})
