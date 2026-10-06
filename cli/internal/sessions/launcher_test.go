package sessions_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/sessions"
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
		}
	})

	work := []sessions.Window{{Name: "api", Dir: "/p/api"}, {Name: "web", Dir: "/p/web"}}

	Describe("Ensure", func() {
		It("only checks existence when the session already exists", func() {
			Expect(l.Ensure("work", work)).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=work"},
			}))
		})

		It("creates a detached session with one window per template window when missing", func() {
			byErr["has-session"] = errors.New("exit status 1")

			Expect(l.Ensure("work", work)).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=work"},
				{"new-session", "-d", "-s", "work", "-n", "api", "-c", "/p/api"},
				{"new-window", "-t", "=work:", "-n", "web", "-c", "/p/web"},
			}))
		})

		It("types each window's command into it", func() {
			byErr["has-session"] = errors.New("exit status 1")

			Expect(l.Ensure("work", []sessions.Window{
				{Name: "code", Dir: "/p/api", Command: "nvim"},
				{Name: "shell", Dir: "/p/api"},
				{Name: "server", Dir: "/p/api", Command: "make dev"},
			})).To(Succeed())

			Expect(calls).To(ContainElements(
				[]string{"send-keys", "-t", "=work:code", "nvim", "Enter"},
				[]string{"send-keys", "-t", "=work:server", "make dev", "Enter"},
			))
			Expect(calls).To(HaveLen(6))
		})

		It("propagates a failure to type a command", func() {
			byErr["has-session"] = errors.New("exit status 1")
			byErr["send-keys"] = errors.New("pane gone")

			Expect(l.Ensure("work", []sessions.Window{{Name: "code", Dir: "/p", Command: "nvim"}})).
				To(MatchError(ContainSubstring("pane gone")))
		})
	})

	Describe("List", func() {
		It("reads the live sessions from list-sessions", func() {
			l.Output = func(args ...string) (string, error) {
				calls = append(calls, args)
				return "api\t3\t1\nscratch\t1\t0\n", nil
			}

			Expect(l.List()).To(Equal([]sessions.Live{
				{Name: "api", Windows: 3, Attached: true},
				{Name: "scratch", Windows: 1},
			}))
			Expect(calls).To(Equal([][]string{
				{"list-sessions", "-F", "#{session_name}\t#{session_windows}\t#{session_attached}"},
			}))
		})

		It("has no sessions when no server is running", func() {
			l.Output = func(...string) (string, error) {
				return "", errors.New("no server running on /tmp/tmux-501/default")
			}

			Expect(l.List()).To(BeEmpty())
		})

		It("has no sessions when the server lists none", func() {
			l.Output = func(...string) (string, error) { return "\n", nil }

			Expect(l.List()).To(BeEmpty())
		})

		It("propagates other failures", func() {
			l.Output = func(...string) (string, error) { return "", errors.New("byobu: not found") }

			_, err := l.List()
			Expect(err).To(MatchError(ContainSubstring("not found")))
		})
	})

	Describe("Kill", func() {
		It("kills the session by exact name", func() {
			Expect(l.Kill("api")).To(Succeed())

			Expect(calls).To(Equal([][]string{{"kill-session", "-t", "=api"}}))
		})
	})

	Describe("Rename", func() {
		It("renames the session by exact name", func() {
			Expect(l.Rename("api", "api-old")).To(Succeed())

			Expect(calls).To(Equal([][]string{{"rename-session", "-t", "=api", "api-old"}}))
		})

		It("refuses names tmux would read as a target", func() {
			Expect(l.Rename("api", "api.v2")).To(MatchError(ContainSubstring("'.' and ':'")))
			Expect(calls).To(BeEmpty())
		})
	})

	Describe("Open", func() {
		It("propagates a session-creation failure", func() {
			byErr["has-session"] = errors.New("exit status 1")
			byErr["new-session"] = errors.New("no server")

			Expect(l.Open("work", work)).To(MatchError(ContainSubstring("no server")))
		})

		It("propagates an attach failure", func() {
			byErr["attach-session"] = errors.New("not a terminal")

			Expect(l.Open("work", work)).To(MatchError(ContainSubstring("not a terminal")))
		})

		It("propagates a window-creation failure", func() {
			byErr["has-session"] = errors.New("exit status 1")
			byErr["new-window"] = errors.New("bad window")

			Expect(l.Open("work", work)).To(MatchError(ContainSubstring("bad window")))
		})

		It("ensures then attaches when outside tmux", func() {
			Expect(l.Open("work", work)).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=work"},
				{"attach-session", "-t", "=work"},
			}))
		})

		It("ensures then switches the current client when inside tmux", func() {
			tmuxEnv = "/tmp/tmux-501/default,123,0"

			Expect(l.Open("work", work)).To(Succeed())

			Expect(calls).To(Equal([][]string{
				{"has-session", "-t", "=work"},
				{"switch-client", "-t", "=work"},
			}))
		})
	})
})
