package tui_test

import (
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/projects"
	"tars/internal/sessions"
	"tars/internal/tui"
)

type fakeManager struct {
	rows      []tui.Row
	rowsErr   error
	log       []string
	err       error
	templates map[string]projects.Template
}

func (m *fakeManager) Template(project string) (projects.Template, error) {
	return m.templates[project], nil
}
func (m *fakeManager) SaveTemplate(project string, t projects.Template) error {
	m.templates[project] = t
	return nil
}

func (m *fakeManager) Rows() ([]tui.Row, error) { return m.rows, m.rowsErr }
func (m *fakeManager) Kill(name string) error {
	m.log = append(m.log, "kill:"+name)
	return m.err
}
func (m *fakeManager) Rename(old, name string) error {
	m.log = append(m.log, "rename:"+old+":"+name)
	return m.err
}

var _ = Describe("Sessions", func() {
	var (
		mgr *fakeManager
		h   *harness[tui.Sessions]
	)

	BeforeEach(func() {
		mgr = &fakeManager{rows: []tui.Row{
			{Name: "api", Live: true, Windows: 3, Attached: true},
			{Name: "infra", Dir: "/p/infra"},
		}}
		h = &harness[tui.Sessions]{model: tui.NewSessions(mgr)}
		h.run(h.model.Init())
		h.send(tea.WindowSizeMsg{Width: 80, Height: 20})
	})

	It("lists running sessions and projects without one", func() {
		Expect(h.view()).To(SatisfyAll(
			ContainSubstring("api"),
			ContainSubstring("3 windows · attached"),
			ContainSubstring("infra"),
			ContainSubstring("not running · /p/infra"),
		))
	})

	It("says once that there is nothing to manage", func() {
		mgr.rows = nil

		h.run(h.model.Init())

		Expect(strings.Count(h.view(), "No items")).To(Equal(1))
	})

	It("shows its keys in the help line", func() {
		Expect(h.view()).To(SatisfyAll(
			ContainSubstring("enter open"),
			ContainSubstring("x kill"),
			ContainSubstring("r rename"),
			ContainSubstring("t edit"),
			ContainSubstring("q quit"),
		))
	})

	It("quits on enter with the selected row to open", func() {
		h.send(tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: tea.KeyEnter})

		Expect(h.quit).To(BeTrue())
		Expect(h.model.Chosen()).To(Equal("infra"))
	})

	It("kills the selected session once confirmed, then reloads the rows", func() {
		h.send(tea.KeyPressMsg{Code: 'x', Text: "x"})
		Expect(h.view()).To(ContainSubstring("Kill session api?"))
		mgr.rows = mgr.rows[1:]

		h.send(tea.KeyPressMsg{Code: 'y', Text: "y"})

		Expect(mgr.log).To(Equal([]string{"kill:api"}))
		Expect(h.view()).NotTo(ContainSubstring("3 windows"))
		Expect(h.quit).To(BeFalse())
	})

	It("closes a cancelled form without acting or quitting", func() {
		h.send(tea.KeyPressMsg{Code: 'x', Text: "x"}, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

		Expect(mgr.log).To(BeEmpty())
		Expect(h.view()).To(ContainSubstring("3 windows"))
		Expect(h.quit).To(BeFalse())
	})

	It("has nothing to kill on a project that is not running", func() {
		h.send(tea.KeyPressMsg{Code: tea.KeyDown}, tea.KeyPressMsg{Code: 'x', Text: "x"})

		Expect(h.view()).NotTo(ContainSubstring("Kill session"))
	})

	It("shows why an action failed and stays open", func() {
		mgr.err = errors.New("no server running")

		h.send(tea.KeyPressMsg{Code: 'x', Text: "x"}, tea.KeyPressMsg{Code: 'y', Text: "y"})

		Expect(h.view()).To(ContainSubstring("no server running"))
		Expect(h.quit).To(BeFalse())
		Expect(lipgloss.Height(h.view())).To(BeNumerically("<=", 20))
	})

	It("shows why the rows could not be loaded", func() {
		mgr.rowsErr = errors.New("byobu: not found")

		h.run(h.model.Init())

		Expect(h.view()).To(ContainSubstring("byobu: not found"))
	})
})

var _ = Describe("Sessions under a real program", func() {
	var (
		mgr *fakeManager
		tm  *teatest.TestModel
	)

	BeforeEach(func() {
		mgr = &fakeManager{
			rows:      []tui.Row{{Name: "api", Live: true, Windows: 3, Dir: "/p/api"}},
			templates: map[string]projects.Template{"api": {Windows: []sessions.Window{{Name: "code", Command: "nvim"}}}},
		}
		tm = teatest.NewTestModel(GinkgoTB(), tui.NewSessions(mgr), teatest.WithInitialTermSize(80, 20))
	})

	// waitFor matches text the renderer writes in one piece; it redraws only changed cells.
	waitFor := func(text string) {
		By("waiting for " + text)
		teatest.WaitFor(GinkgoTB(), tm.Output(), func(b []byte) bool {
			return strings.Contains(ansi.Strip(string(b)), text)
		}, teatest.WithDuration(5*time.Second))
	}

	It("types action keys into the filter instead of acting on them", func() {
		waitFor("3 windows")
		tm.Send(tea.KeyPressMsg{Code: '/', Text: "/"})
		tm.Send(tea.KeyPressMsg{Code: 'x', Text: "x"})
		tm.Send(tea.KeyPressMsg{Code: 'y', Text: "y"})
		waitFor("xy")
		tm.Send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

		tm.WaitFinished(GinkgoTB(), teatest.WithFinalTimeout(5*time.Second))
		Expect(mgr.log).To(BeEmpty())
	})

	It("edits the selected project's template and saves it", func() {
		waitFor("3 windows")
		tm.Send(tea.KeyPressMsg{Code: 't', Text: "t"})
		waitFor("template for api")
		for _, r := range "logs | log" {
			tm.Send(tea.KeyPressMsg{Code: r, Text: string(r)})
		}
		tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
		waitFor("byobu")
		tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})

		tm.WaitFinished(GinkgoTB(), teatest.WithFinalTimeout(5*time.Second))
		Expect(mgr.templates["api"]).To(Equal(projects.Template{Windows: []sessions.Window{
			{Name: "code", Command: "nvim"},
			{Name: "logs", Dir: "log"},
		}}))
	})

	It("renames a session typed into the form, then quits without opening anything", func() {
		waitFor("3 windows")
		tm.Send(tea.KeyPressMsg{Code: 'r', Text: "r"})
		waitFor("Rename")
		tm.Send(tea.KeyPressMsg{Code: '2', Text: "2"})
		tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
		waitFor("byobu")
		tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})

		final := tm.FinalModel(GinkgoTB(), teatest.WithFinalTimeout(5*time.Second)).(tui.Sessions)
		Expect(mgr.log).To(Equal([]string{"rename:api:api2"}))
		Expect(final.Chosen()).To(BeEmpty())
	})
})
