package cmd_test

import (
	"bytes"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/projects"
	"tars/internal/sessions"
	"tars/internal/tui"
)

type spyLive struct {
	live []sessions.Live
	log  []string
}

func (l *spyLive) List() ([]sessions.Live, error) { return l.live, nil }
func (l *spyLive) Attach(name string) error {
	l.log = append(l.log, "attach:"+name)
	return nil
}
func (l *spyLive) Kill(name string) error {
	l.log = append(l.log, "kill:"+name)
	return nil
}
func (l *spyLive) Rename(old, name string) error {
	l.log = append(l.log, "rename:"+old+":"+name)
	return nil
}
func (l *spyLive) Open(name string, windows []sessions.Window) error {
	var parts []string
	for _, w := range windows {
		parts = append(parts, w.Name+"@"+w.Dir)
	}
	l.log = append(l.log, "open:"+name+":"+strings.Join(parts, ","))
	return nil
}

type spyCatalog struct {
	projects  []projects.Project
	templates projects.File
	seeded    bool
}

func (c *spyCatalog) Projects() ([]projects.Project, error) { return c.projects, nil }
func (c *spyCatalog) Templates() (projects.File, error)     { return c.templates, nil }
func (c *spyCatalog) Path() string                          { return "/cfg/projects.yaml" }
func (c *spyCatalog) SaveTemplates(f projects.File) error {
	c.templates = f
	return nil
}
func (c *spyCatalog) Seed() error {
	c.seeded = true
	return nil
}

type spySessionPicker struct {
	options []string
	choice  string
	err     error
}

func (p *spySessionPicker) Pick(options []string) (string, error) {
	p.options = options
	return p.choice, p.err
}

type sessionsFixture struct {
	live    *spyLive
	catalog *spyCatalog
	stdout  *bytes.Buffer
	s       *cmd.Sessions
}

func newSessionsFixture() *sessionsFixture {
	f := &sessionsFixture{
		live: &spyLive{live: []sessions.Live{{Name: "scratch", Windows: 1}}},
		catalog: &spyCatalog{
			projects: []projects.Project{{Name: "api", Dir: "/p/api"}, {Name: "infra", Dir: "/p/infra"}},
			templates: projects.File{Projects: map[string]projects.Template{"api": {Windows: []sessions.Window{
				{Name: "code"}, {Name: "web", Dir: "frontend"},
			}}}},
		},
		stdout: &bytes.Buffer{},
	}
	f.s = &cmd.Sessions{Live: f.live, Catalog: f.catalog, Stdout: f.stdout}
	return f
}

var _ = Describe("Sessions.List", func() {
	It("prints each running session, marking the attached one and its project", func() {
		f := newSessionsFixture()
		f.live.live = []sessions.Live{{Name: "api", Windows: 3, Attached: true}, {Name: "scratch", Windows: 1}}

		Expect(f.s.List()).To(Succeed())

		Expect(strings.Split(f.stdout.String(), "\n")).To(HaveExactElements(
			MatchRegexp(`^\* api\s+3 windows\s+project /p/api$`),
			MatchRegexp(`^  scratch\s+1 windows$`),
			"",
		))
	})
})

var _ = Describe("Sessions.Projects", func() {
	It("prints each project, marking running ones and whether it has its own template", func() {
		f := newSessionsFixture()
		f.live.live = []sessions.Live{{Name: "infra", Windows: 1}}

		Expect(f.s.Projects()).To(Succeed())

		Expect(strings.Split(f.stdout.String(), "\n")).To(HaveExactElements(
			MatchRegexp(`^  api\s+/p/api\s+template$`),
			MatchRegexp(`^● infra\s+/p/infra\s+default$`),
			"",
		))
	})
})

var _ = Describe("Sessions.Kill", func() {
	It("kills the running session", func() {
		f := newSessionsFixture()

		Expect(f.s.Kill("scratch")).To(Succeed())

		Expect(f.live.log).To(Equal([]string{"kill:scratch"}))
	})
})

var _ = Describe("Sessions.Rename", func() {
	It("renames the running session", func() {
		f := newSessionsFixture()

		Expect(f.s.Rename("scratch", "notes")).To(Succeed())

		Expect(f.live.log).To(Equal([]string{"rename:scratch:notes"}))
	})
})

var _ = Describe("Sessions.Edit", func() {
	It("seeds the projects file, then opens it in the editor", func() {
		f := newSessionsFixture()
		var edited string
		f.s.EditFn = func(path string) error {
			edited = path
			return nil
		}

		Expect(f.s.Edit()).To(Succeed())

		Expect(f.catalog.seeded).To(BeTrue())
		Expect(edited).To(Equal("/cfg/projects.yaml"))
	})
})

var _ = Describe("Sessions.PickAndOpen", func() {
	It("offers running sessions, then projects without one, and starts the chosen project", func() {
		f := newSessionsFixture()
		f.live.live = []sessions.Live{{Name: "scratch"}, {Name: "infra"}}
		picker := &spySessionPicker{choice: "api (new)"}
		f.s.Picker = picker

		Expect(f.s.PickAndOpen()).To(Succeed())

		Expect(picker.options).To(Equal([]string{"scratch", "infra", "api (new)"}))
		Expect(f.live.log).To(Equal([]string{"open:api:code@/p/api,web@/p/api/frontend"}))
	})

	It("treats a user-aborted picker as a quiet no-op", func() {
		f := newSessionsFixture()
		f.s.Picker = &spySessionPicker{err: errors.New("user aborted")}

		Expect(f.s.PickAndOpen()).To(Succeed())
		Expect(f.live.log).To(BeEmpty())
	})
})

var _ = Describe("Sessions.Interactive", func() {
	var (
		f       *sessionsFixture
		manager tui.Manager
	)

	BeforeEach(func() {
		f = newSessionsFixture()
		f.live.live = []sessions.Live{{Name: "api", Windows: 2, Attached: true}, {Name: "scratch", Windows: 1}}
		f.s.Manage = func(m tui.Manager) (string, error) {
			manager = m
			return "", nil
		}
		Expect(f.s.Interactive()).To(Succeed())
	})

	It("manages the running sessions, then the projects without one", func() {
		Expect(manager.Rows()).To(Equal([]tui.Row{
			{Name: "api", Live: true, Windows: 2, Attached: true, Dir: "/p/api"},
			{Name: "scratch", Live: true, Windows: 1},
			{Name: "infra", Dir: "/p/infra"},
		}))
	})

	It("kills and renames running sessions", func() {
		Expect(manager.Kill("scratch")).To(Succeed())
		Expect(manager.Rename("api", "api-old")).To(Succeed())

		Expect(f.live.log).To(Equal([]string{"kill:scratch", "rename:api:api-old"}))
	})

	It("edits a project's configured or default template", func() {
		Expect(manager.Template("api")).To(Equal(projects.Template{Windows: []sessions.Window{
			{Name: "code"}, {Name: "web", Dir: "frontend"},
		}}))
		Expect(manager.Template("infra")).To(Equal(projects.Template{Windows: []sessions.Window{{Name: "infra"}}}))
	})

	It("saves an edited template alongside the others", func() {
		infra := projects.Template{Windows: []sessions.Window{{Name: "tf", Command: "terraform plan"}}}

		Expect(manager.SaveTemplate("infra", infra)).To(Succeed())

		Expect(f.catalog.templates.Projects).To(HaveKeyWithValue("infra", infra))
		Expect(f.catalog.templates.Projects).To(HaveKey("api"))
	})

	It("lists the running sessions instead when headless", func() {
		f.s.Headless = true
		f.s.Manage = func(tui.Manager) (string, error) { panic("no terminal to manage in") }

		Expect(f.s.Interactive()).To(Succeed())

		Expect(f.stdout.String()).To(ContainSubstring("* api"))
	})

	It("opens what was picked once the manager has exited", func() {
		f.s.Manage = func(tui.Manager) (string, error) { return "infra", nil }

		Expect(f.s.Interactive()).To(Succeed())

		Expect(f.live.log).To(Equal([]string{"open:infra:infra@/p/infra"}))
	})
})

var _ = Describe("Sessions.Open", func() {
	It("attaches to a running session by name", func() {
		f := newSessionsFixture()

		Expect(f.s.Open("scratch")).To(Succeed())

		Expect(f.live.log).To(Equal([]string{"attach:scratch"}))
	})

	It("starts a project's session from its template when none is running", func() {
		f := newSessionsFixture()

		Expect(f.s.Open("api")).To(Succeed())

		Expect(f.live.log).To(Equal([]string{"open:api:code@/p/api,web@/p/api/frontend"}))
	})

	It("opens the project containing the current dir when no name is given", func() {
		f := newSessionsFixture()
		f.s.Getwd = func() (string, error) { return "/p/infra/modules", nil }

		Expect(f.s.Open("")).To(Succeed())

		Expect(f.live.log).To(Equal([]string{"open:infra:infra@/p/infra"}))
	})

	It("errors when no name is given outside every project", func() {
		f := newSessionsFixture()
		f.s.Getwd = func() (string, error) { return "/p/api2", nil }

		Expect(f.s.Open("")).To(MatchError(ContainSubstring("/p/api2 is not inside a project")))
		Expect(f.live.log).To(BeEmpty())
	})

	It("errors on an unknown name, listing what can be opened", func() {
		f := newSessionsFixture()

		Expect(f.s.Open("nope")).To(MatchError(And(
			ContainSubstring(`"nope"`),
			ContainSubstring("scratch"),
			ContainSubstring("api, infra"),
		)))
		Expect(f.live.log).To(BeEmpty())
	})
})
