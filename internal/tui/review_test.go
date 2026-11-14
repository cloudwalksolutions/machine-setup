package tui_test

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/aymanbagabas/go-udiff"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
	"tars/internal/fsutil"
	"tars/internal/tui"
)

var _ = Describe("Review", func() {
	var h *harness[tui.Review]

	diffOf := func(path, old, new string) udiff.UnifiedDiff {
		d, err := udiff.ToUnifiedDiff(path, path, old, udiff.Lines(old, new), 3)
		Expect(err).NotTo(HaveOccurred())
		return d
	}
	plans := func() []components.Plan {
		return []components.Plan{
			{Component: "zsh", Changes: []fsutil.Change{
				{Path: "/h/.zshrc", Kind: fsutil.Overwrite, Backup: "v2", Diff: diffOf("/h/.zshrc", "keep\nold\n", "keep\nnew\n")},
				{Path: "/h/.zshrc_aliases", Kind: fsutil.Create},
			}},
			{Component: "fonts", Changes: []fsutil.Change{
				{Path: "/h/Hack.ttf", Kind: fsutil.Overwrite, Backup: "v1", Binary: true},
			}},
		}
	}

	BeforeEach(func() {
		h = &harness[tui.Review]{model: tui.NewReview(plans())}
		h.send(tea.WindowSizeMsg{Width: 120, Height: 40})
	})

	It("lists every component with its files and marks, shows the first file's diff, and totals the plan", func() {
		view := h.view()

		Expect(view).To(SatisfyAll(
			ContainSubstring("✓ zsh  1 create, 1 modify"),
			ContainSubstring("~ /h/.zshrc  +1 −1"),
			ContainSubstring("+ /h/.zshrc_aliases"),
			ContainSubstring("✓ fonts  1 modify"),
			ContainSubstring("~ /h/Hack.ttf  binary"),
			ContainSubstring("-old"),
			ContainSubstring("+new"),
			ContainSubstring("Plan: 1 to create, 2 to modify · 2 components"),
			ContainSubstring("2 of 2 accepted"),
		))
	})

	It("moves the pointer down and up and shows the selected change on the right", func() {
		Expect(h.view()).To(ContainSubstring("▶ ✓ zsh"))

		h.send(tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: 'j', Text: "j"})

		Expect(h.view()).To(SatisfyAll(ContainSubstring("▶   + /h/.zshrc_aliases"), ContainSubstring("/h/.zshrc_aliases · create"), ContainSubstring("new file")))
		Expect(h.view()).NotTo(SatisfyAny(ContainSubstring("-old"), ContainSubstring("▶ ✓ zsh")))

		h.send(tea.KeyPressMsg{Code: 'k', Text: "k"})

		Expect(h.view()).To(ContainSubstring("/h/.zshrc · modify · +1 −1"))
	})

	It("toggles the selected row's component with space, from a file row too, and recounts", func() {
		h.send(tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

		Expect(h.view()).To(SatisfyAll(ContainSubstring("✗ zsh"), ContainSubstring("1 of 2 accepted")))
		Expect(h.model.Accepted()).To(Equal([]string{"fonts"}))

		h.send(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

		Expect(h.model.Accepted()).To(Equal([]string{"zsh", "fonts"}))
	})

	It("skips everything with d and accepts everything with a", func() {
		h.send(tea.KeyPressMsg{Code: 'd', Text: "d"})

		Expect(h.view()).To(ContainSubstring("0 of 2 accepted"))
		Expect(h.model.Accepted()).To(BeEmpty())

		h.send(tea.KeyPressMsg{Code: 'a', Text: "a"})

		Expect(h.model.Accepted()).To(Equal([]string{"zsh", "fonts"}))
	})

	It("jumps between hunks with n/N and between files with ]/[", func() {
		old := "a1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\na12\n"
		new := "b1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\nb12\n"
		h = &harness[tui.Review]{model: tui.NewReview([]components.Plan{
			{Component: "zsh", Changes: []fsutil.Change{
				{Path: "/h/.zshrc", Kind: fsutil.Overwrite, Diff: diffOf("/h/.zshrc", old, new)},
				{Path: "/h/.zshrc_aliases", Kind: fsutil.Create},
			}},
		})}
		h.send(tea.WindowSizeMsg{Width: 120, Height: 12})
		Expect(h.view()).To(SatisfyAll(ContainSubstring("2 hunks"), ContainSubstring("-a1"), Not(ContainSubstring("-a12"))))

		h.send(tea.KeyPressMsg{Code: 'n', Text: "n"})
		Expect(h.view()).To(SatisfyAll(ContainSubstring("-a12"), Not(ContainSubstring("-a1\n"))))

		h.send(tea.KeyPressMsg{Code: 'N', Text: "N"})
		Expect(h.view()).To(ContainSubstring("-a1"))

		h.send(tea.KeyPressMsg{Code: ']', Text: "]"})
		Expect(h.view()).To(ContainSubstring("/h/.zshrc_aliases · create"))

		h.send(tea.KeyPressMsg{Code: '[', Text: "["})
		Expect(h.view()).To(ContainSubstring("/h/.zshrc · modify"))
	})

	It("scrolls the diff instead of moving the selection once tab focuses the right pane", func() {
		old := "a1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\na12\n"
		new := "b1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\nb12\n"
		h = &harness[tui.Review]{model: tui.NewReview([]components.Plan{
			{Component: "zsh", Changes: []fsutil.Change{
				{Path: "/h/.zshrc", Kind: fsutil.Overwrite, Diff: diffOf("/h/.zshrc", old, new)},
				{Path: "/h/.zshrc_aliases", Kind: fsutil.Create},
			}},
		})}
		h.send(tea.WindowSizeMsg{Width: 120, Height: 12}, tea.KeyPressMsg{Code: tea.KeyTab})
		for range 3 {
			h.send(tea.KeyPressMsg{Code: 'j', Text: "j"})
		}

		Expect(h.view()).To(SatisfyAll(
			ContainSubstring("@@ -9,4 +9,4 @@"),
			Not(ContainSubstring("--- /h/.zshrc")),
			ContainSubstring("/h/.zshrc · modify"),
		))

		h.send(tea.KeyPressMsg{Code: tea.KeyTab}, tea.KeyPressMsg{Code: 'j', Text: "j"}, tea.KeyPressMsg{Code: 'j', Text: "j"})

		Expect(h.view()).To(ContainSubstring("/h/.zshrc_aliases · create"))
	})

	It("scrolls the diff with the wheel and selects a row on click", func() {
		old := "a1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\na12\n"
		new := "b1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\nb12\n"
		h = &harness[tui.Review]{model: tui.NewReview([]components.Plan{
			{Component: "zsh", Changes: []fsutil.Change{
				{Path: "/h/.zshrc", Kind: fsutil.Overwrite, Diff: diffOf("/h/.zshrc", old, new)},
				{Path: "/h/.zshrc_aliases", Kind: fsutil.Create},
			}},
		})}
		h.send(tea.WindowSizeMsg{Width: 120, Height: 12}, tea.MouseWheelMsg{Button: tea.MouseWheelDown})

		Expect(h.view()).NotTo(ContainSubstring("--- /h/.zshrc"))

		h.run(h.model.View().OnMouse(tea.MouseClickMsg{X: 4, Y: 3, Button: tea.MouseLeft}))

		Expect(h.view()).To(ContainSubstring("/h/.zshrc_aliases · create"))
	})

	It("pages and jumps through the diff with pgdn/pgup and G/g from either pane", func() {
		old := "a1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\na12\n"
		new := "b1\na2\na3\na4\na5\na6\na7\na8\na9\na10\na11\nb12\n"
		h = &harness[tui.Review]{model: tui.NewReview([]components.Plan{
			{Component: "zsh", Changes: []fsutil.Change{{Path: "/h/.zshrc", Kind: fsutil.Overwrite, Diff: diffOf("/h/.zshrc", old, new)}}},
		})}
		h.send(tea.WindowSizeMsg{Width: 120, Height: 12}, tea.KeyPressMsg{Code: 'G', Text: "G"})
		Expect(h.view()).To(SatisfyAll(ContainSubstring("+b12"), Not(ContainSubstring("--- /h/.zshrc"))))

		h.send(tea.KeyPressMsg{Code: 'g', Text: "g"})
		Expect(h.view()).To(ContainSubstring("--- /h/.zshrc"))

		h.send(tea.KeyPressMsg{Code: tea.KeyPgDown})
		Expect(h.view()).NotTo(ContainSubstring("--- /h/.zshrc"))

		h.send(tea.KeyPressMsg{Code: tea.KeyPgUp})
		Expect(h.view()).To(ContainSubstring("--- /h/.zshrc"))
	})

	It("fits the terminal, and says so when it is too small to split", func() {
		for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 20}, {Width: 60, Height: 12}} {
			h.send(size)
			Expect(lipgloss.Height(h.view())).To(BeNumerically("<=", size.Height), "%dx%d", size.Width, size.Height)
			Expect(lipgloss.Width(h.view())).To(BeNumerically("<=", size.Width), "%dx%d", size.Width, size.Height)
		}

		h.send(tea.WindowSizeMsg{Width: 40, Height: 8})

		Expect(h.view()).To(ContainSubstring("terminal too small"))
		Expect(lipgloss.Height(h.view())).To(BeNumerically("<=", 8))
	})

	It("runs under a real program: toggle a component, apply, and hand back the accepted set", func() {
		tm := teatest.NewTestModel(GinkgoTB(), tui.NewReview(plans()), teatest.WithInitialTermSize(120, 40))
		teatest.WaitFor(GinkgoTB(), tm.Output(), func(b []byte) bool {
			return strings.Contains(ansi.Strip(string(b)), "2 of 2 accepted")
		}, teatest.WithDuration(5*time.Second))

		tm.Send(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
		tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

		final := tm.FinalModel(GinkgoTB(), teatest.WithFinalTimeout(5*time.Second)).(tui.Review)
		Expect(final.Aborted()).To(BeFalse())
		Expect(final.Accepted()).To(Equal([]string{"fonts"}))
	})

	It("applies on enter and aborts on q or ctrl+c, quitting either way", func() {
		h.send(tea.KeyPressMsg{Code: 'd', Text: "d"}, tea.KeyPressMsg{Code: tea.KeyEnter})

		Expect(h.quit).To(BeTrue())
		Expect(h.model.Aborted()).To(BeFalse())

		for _, quit := range []tea.KeyPressMsg{{Code: 'q', Text: "q"}, {Code: 'c', Mod: tea.ModCtrl}} {
			h = &harness[tui.Review]{model: tui.NewReview(plans())}
			h.send(tea.WindowSizeMsg{Width: 120, Height: 40}, quit)

			Expect(h.quit).To(BeTrue())
			Expect(h.model.Aborted()).To(BeTrue())
		}
	})
})
