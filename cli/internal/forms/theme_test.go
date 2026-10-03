package forms

import (
	"strings"

	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func lineContaining(view, needle string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(ansi.Strip(line), needle) {
			return line
		}
	}
	Fail("no line contains " + needle)
	return ""
}

var _ = Describe("theme", func() {
	var view string

	BeforeEach(func() {
		var first, second bool
		f := huh.NewForm(huh.NewGroup(
			confirm("First question", "", &first),
			confirm("Second question", "", &second),
		)).WithTheme(theme()).WithKeyMap(keymap())
		_ = f.Init()
		view = f.View()
	})

	It("points at the focused field's title only", func() {
		Expect(ansi.Strip(view)).To(ContainSubstring("▶ First question"))
		Expect(ansi.Strip(view)).NotTo(ContainSubstring("▶ Second question"))
	})

	It("dims the blurred field's title", func() {
		Expect(lineContaining(view, "Second question")).To(ContainSubstring("38;5;240m"))
		Expect(lineContaining(view, "First question")).NotTo(ContainSubstring("38;5;240m"))
	})

	It("left-aligns confirm buttons under their title", func() {
		title := ansi.Strip(lineContaining(view, "First question"))
		buttons := ansi.Strip(lineContaining(view, "Yes"))
		Expect(strings.Index(buttons, "Yes")).To(BeNumerically("<=", strings.Index(title, "First")))
	})

	It("toggles multi-select rows with space only", func() {
		Expect(keymap().MultiSelect.Toggle.Keys()).To(Equal([]string{"space"}))
	})
})
