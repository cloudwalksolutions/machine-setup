package sessions_test

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/sessions"
)

var _ = Describe("Launcher", func() {
	var (
		calls   [][]string
		byErr   map[string]error // command name -> error to return
		tmuxEnv string
		l       sessions.Launcher
	)

	BeforeEach(func() {
		calls = nil
		byErr = map[string]error{}
		tmuxEnv = ""
		l = sessions.Launcher{
			Run: func(args ...string) error {
				calls = append(calls, args)
				return byErr[args[0]]
			},
			LookupEnv: func(key string) (string, bool) {
				if key == "TMUX" && tmuxEnv != "" {
					return tmuxEnv, true
				}
				return "", false
			},
			Stdout: GinkgoWriter,
			Stderr: GinkgoWriter,
		}
	})

	work := sessions.Session{Name: "work", Dirs: []string{"/p/api", "/p/web"}}

	Describe("Ensure", func() {
		It("only checks existence when the session already exists", func() {
			Expect(l.Ensure(work)).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=work"},
			}))
		})

		It("creates a detached session with one window per dir when missing", func() {
			byErr["has-session"] = errors.New("exit status 1")

			Expect(l.Ensure(work)).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=work"},
				{"new-session", "-d", "-s", "work", "-n", "api", "-c", "/p/api"},
				{"new-window", "-t", "=work:", "-n", "web", "-c", "/p/web"},
			}))
		})
	})

	Describe("Fresh", func() {
		It("opens a named session rooted at home", func() {
			byErr["has-session"] = errors.New("exit status 1")
			l.Home = "/fake/home"

			Expect(l.Fresh("scratch")).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=scratch"},
				{"new-session", "-d", "-s", "scratch", "-n", "home", "-c", "/fake/home"},
				{"attach-session", "-t", "=scratch"},
			}))
		})

		It("opens an attached unnamed session when no name is given outside tmux", func() {
			Expect(l.Fresh("")).To(Succeed())

			Expect(calls).To(Equal([][]string{{"new-session"}}))
		})

		It("refuses an unnamed session inside tmux, asking for a name", func() {
			tmuxEnv = "/tmp/tmux-501/default,123,0"

			Expect(l.Fresh("")).To(MatchError(ContainSubstring("name")))
			Expect(calls).To(BeEmpty())
		})
	})

	Describe("OpenAll", func() {
		It("ensures every session, reports failures, and attaches to the first that succeeded", func() {
			var stderr strings.Builder
			l.Stderr = &stderr
			l.Run = func(args ...string) error {
				calls = append(calls, args)
				if args[0] == "has-session" {
					return errors.New("exit status 1")
				}
				if args[0] == "new-session" && args[3] == "bad" {
					return errors.New("boom")
				}
				return nil
			}
			f := sessions.File{Sessions: []sessions.Session{
				{Name: "bad", Dirs: []string{"/b"}},
				{Name: "good", Dirs: []string{"/g"}},
			}}

			Expect(l.OpenAll(f)).To(Succeed())

			Expect(stderr.String()).To(ContainSubstring("bad"))
			Expect(calls).To(ContainElement([]string{"new-session", "-d", "-s", "good", "-n", "g", "-c", "/g"}))
			Expect(calls[len(calls)-1]).To(Equal([]string{"attach-session", "-t", "=good"}))
		})

		It("errors when no session could be ensured", func() {
			l.Run = func(args ...string) error { return errors.New("boom") }

			Expect(l.OpenAll(sessions.File{Sessions: []sessions.Session{
				{Name: "only", Dirs: []string{"/o"}},
			}})).NotTo(Succeed())
		})
	})

	Describe("Open", func() {
		It("ensures then attaches when outside tmux", func() {
			Expect(l.Open(work)).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=work"},
				{"attach-session", "-t", "=work"},
			}))
		})

		It("ensures then switches the current client when inside tmux", func() {
			tmuxEnv = "/tmp/tmux-501/default,123,0"

			Expect(l.Open(work)).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=work"},
				{"switch-client", "-t", "=work"},
			}))
		})
	})
})
