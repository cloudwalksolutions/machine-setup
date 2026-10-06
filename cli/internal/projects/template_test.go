package projects_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/projects"
	"tars/internal/sessions"
)

var _ = Describe("ParseTemplate", func() {
	It("reads one `name | dir | command` window per line", func() {
		Expect(projects.ParseTemplate("code | | nvim\nweb | frontend | make dev\n")).To(Equal(projects.Template{
			Windows: []sessions.Window{
				{Name: "code", Command: "nvim"},
				{Name: "web", Dir: "frontend", Command: "make dev"},
			},
		}))
	})

	It("keeps pipes inside a command", func() {
		Expect(projects.ParseTemplate("logs | | make dev | tee dev.log")).To(Equal(projects.Template{
			Windows: []sessions.Window{{Name: "logs", Command: "make dev | tee dev.log"}},
		}))
	})

	It("skips blank lines and lets dir and command be left off", func() {
		Expect(projects.ParseTemplate("\nshell\n  \nlogs | log\n")).To(Equal(projects.Template{
			Windows: []sessions.Window{{Name: "shell"}, {Name: "logs", Dir: "log"}},
		}))
	})

	DescribeTable("rejects windows tmux cannot target",
		func(text, problem string) {
			_, err := projects.ParseTemplate(text)
			Expect(err).To(MatchError(ContainSubstring(problem)))
		},
		Entry("no name", "code\n | src | nvim\n", `line 2: window needs a name`),
		Entry("a target separator", "api.v2\n", `line 1: window "api.v2": '.' and ':' are not allowed`),
	)
})

var _ = Describe("Template.String", func() {
	It("writes the lines ParseTemplate reads back", func() {
		t := projects.Template{Windows: []sessions.Window{
			{Name: "code", Command: "nvim"},
			{Name: "web", Dir: "frontend", Command: "make dev"},
			{Name: "shell"},
		}}

		Expect(t.String()).To(Equal("code |  | nvim\nweb | frontend | make dev\nshell |  | \n"))
		Expect(projects.ParseTemplate(t.String())).To(Equal(t))
	})
})

var _ = Describe("File.Template", func() {
	It("is the default one-window template for a project without one", func() {
		api := projects.Project{Name: "api", Dir: "/p/api.v2"}

		Expect(projects.File{}.Template(api)).To(Equal(projects.Template{Windows: []sessions.Window{{Name: "api_v2"}}}))
	})

	It("is the configured template, even with a single window", func() {
		t := projects.Template{Windows: []sessions.Window{{Name: "code", Command: "nvim"}}}

		Expect(projects.File{Projects: map[string]projects.Template{"api": t}}.Template(projects.Project{Name: "api"})).To(Equal(t))
	})
})

var _ = Describe("Windows", func() {
	api := projects.Project{Name: "api", Dir: "/p/api"}

	It("defaults to one window at the project root", func() {
		Expect(projects.File{}.Windows(api)).To(Equal([]sessions.Window{{Name: "api", Dir: "/p/api"}}))
	})

	It("resolves a configured template's dirs against the project", func() {
		f := projects.File{Projects: map[string]projects.Template{"api": {Windows: []sessions.Window{
			{Name: "code", Command: "nvim"},
			{Name: "web", Dir: "frontend", Command: "make dev"},
		}}}}

		Expect(f.Windows(api)).To(Equal([]sessions.Window{
			{Name: "code", Dir: "/p/api", Command: "nvim"},
			{Name: "web", Dir: "/p/api/frontend", Command: "make dev"},
		}))
	})
})
