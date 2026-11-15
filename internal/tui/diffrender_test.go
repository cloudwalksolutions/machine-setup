package tui_test

import (
	"strings"

	"github.com/aymanbagabas/go-udiff"
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/tui"
)

var _ = Describe("RenderDiff", func() {
	unified := func(old, new string) udiff.UnifiedDiff {
		d, err := udiff.ToUnifiedDiff("~/.zshrc", "~/.zshrc", old, udiff.Lines(old, new), 3)
		Expect(err).NotTo(HaveOccurred())
		return d
	}
	plain := func(lines []string) []string {
		out := make([]string, len(lines))
		for i, l := range lines {
			out[i] = strings.TrimRight(ansi.Strip(l), " ")
		}
		return out
	}

	It("renders one line per diff line, keeping the unified text, and indexes where each hunk starts", func() {
		lines, hunks := tui.RenderDiff(unified("keep\nold\n", "keep\nnew\n"), 40, true)

		Expect(plain(lines)).To(Equal([]string{
			"--- ~/.zshrc",
			"+++ ~/.zshrc",
			"@@ -1,2 +1,2 @@",
			" keep",
			"-old",
			"+new",
		}))
		Expect(hunks).To(Equal([]int{2}))
	})

	It("fills removed and added rows with a dark or light tint across the whole width", func() {
		dark, _ := tui.RenderDiff(unified("old\n", "new\n"), 20, true)
		light, _ := tui.RenderDiff(unified("old\n", "new\n"), 20, false)

		Expect(dark[3]).To(ContainSubstring("48;2;58;28;37m-old" + strings.Repeat(" ", 16) + "\x1b[m"))
		Expect(dark[4]).To(ContainSubstring("48;2;28;47;28m"))
		Expect(light[3]).To(ContainSubstring("48;2;251;224;230m"))
		Expect(light[4]).To(ContainSubstring("48;2;214;245;214m"))
	})

	It("emphasizes the words that changed within a line pair, unless most of the line changed", func() {
		rows, _ := tui.RenderDiff(unified("keep this line\nalpha\n", "keep that line\nomega\n"), 40, true)

		Expect(rows[3]).To(ContainSubstring("48;2;107;42;60mthis"))
		Expect(rows[4]).NotTo(ContainSubstring("48;2;107;42;60m"))
		Expect(rows[5]).To(ContainSubstring("48;2;47;90;47mthat"))
		Expect(rows[6]).NotTo(ContainSubstring("48;2;47;90;47m"))
	})
})
