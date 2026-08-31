package cmd_test

import (
	"bytes"
	"errors"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/cmd"
	"github.com/cloudwalk/machine-setup/internal/sessions"
)

type spySessionStore struct {
	file   sessions.File
	err    error
	path   string
	seeded bool
}

func (s *spySessionStore) Load() (sessions.File, error) { return s.file, s.err }
func (s *spySessionStore) Path() string                 { return s.path }
func (s *spySessionStore) Seed() error {
	s.seeded = true
	return nil
}

type spySessionOpener struct {
	log     *[]string
	openErr error
}

func (o *spySessionOpener) Open(s sessions.Session) error {
	*o.log = append(*o.log, "open:"+s.Name)
	return o.openErr
}
func (o *spySessionOpener) OpenAll(f sessions.File) error {
	*o.log = append(*o.log, "all")
	return nil
}
func (o *spySessionOpener) Fresh(name string) error {
	*o.log = append(*o.log, "fresh:"+name)
	return nil
}

type sessionsFixture struct {
	store  *spySessionStore
	opener *spySessionOpener
	log    []string
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	s      *cmd.Sessions
}

func newSessionsFixture() *sessionsFixture {
	f := &sessionsFixture{
		store: &spySessionStore{
			path: "/cfg/sessions.yaml",
			file: sessions.File{Sessions: []sessions.Session{
				{Name: "work", Dirs: []string{"/p/api"}},
				{Name: "personal", Dirs: []string{"/dotfiles"}},
			}},
		},
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	f.opener = &spySessionOpener{log: &f.log}
	f.s = &cmd.Sessions{
		Store:  f.store,
		Opener: f.opener,
		Stdout: f.stdout,
		Stderr: f.stderr,
	}
	return f
}

var _ = Describe("Sessions.Open", func() {
	It("opens the configured session matching the name", func() {
		f := newSessionsFixture()

		Expect(f.s.Open("personal")).To(Succeed())

		Expect(f.log).To(Equal([]string{"open:personal"}))
	})

	It("errors on an unknown name, listing the configured sessions", func() {
		f := newSessionsFixture()

		err := f.s.Open("nope")

		Expect(err).To(MatchError(And(
			ContainSubstring("nope"),
			ContainSubstring("work"),
			ContainSubstring("personal"),
		)))
		Expect(f.log).To(BeEmpty())
	})

	It("suggests `tars sessions edit` when the config file is missing", func() {
		f := newSessionsFixture()
		f.store.err = os.ErrNotExist

		Expect(f.s.Open("work")).To(MatchError(ContainSubstring("tars sessions edit")))
	})
})

var _ = Describe("Sessions.All", func() {
	It("opens every configured session via the opener", func() {
		f := newSessionsFixture()

		Expect(f.s.All()).To(Succeed())

		Expect(f.log).To(Equal([]string{"all"}))
	})
})

var _ = Describe("Sessions.New", func() {
	It("opens a fresh session with the given name, no config needed", func() {
		f := newSessionsFixture()
		f.store.err = os.ErrNotExist

		Expect(f.s.New("scratch")).To(Succeed())

		Expect(f.log).To(Equal([]string{"fresh:scratch"}))
	})
})

type spySessionPicker struct {
	options []string
	choice  string
	err     error
}

func (p *spySessionPicker) Pick(options []string) (string, error) {
	p.options = options
	return p.choice, p.err
}

var _ = Describe("Sessions.PickAndRun", func() {
	It("offers the configured sessions plus all/fresh and opens the chosen one", func() {
		f := newSessionsFixture()
		picker := &spySessionPicker{choice: "personal"}
		f.s.Picker = picker

		Expect(f.s.PickAndRun()).To(Succeed())

		Expect(picker.options).To(Equal([]string{"work", "personal", "(all)", "(fresh)"}))
		Expect(f.log).To(Equal([]string{"open:personal"}))
	})

	It("routes (all) to the opener's OpenAll and (fresh) to an unnamed fresh session", func() {
		f := newSessionsFixture()
		f.s.Picker = &spySessionPicker{choice: "(all)"}
		Expect(f.s.PickAndRun()).To(Succeed())

		f.s.Picker = &spySessionPicker{choice: "(fresh)"}
		Expect(f.s.PickAndRun()).To(Succeed())

		Expect(f.log).To(Equal([]string{"all", "fresh:"}))
	})

	It("treats a user-aborted picker as a quiet no-op", func() {
		f := newSessionsFixture()
		f.s.Picker = &spySessionPicker{err: errors.New("user aborted")}

		Expect(f.s.PickAndRun()).To(Succeed())
		Expect(f.log).To(BeEmpty())
	})
})

var _ = Describe("Sessions.Edit", func() {
	It("seeds the config if missing, then opens it in the editor", func() {
		f := newSessionsFixture()
		var edited string
		f.s.EditFn = func(path string) error {
			edited = path
			return nil
		}

		Expect(f.s.Edit()).To(Succeed())

		Expect(f.store.seeded).To(BeTrue())
		Expect(edited).To(Equal("/cfg/sessions.yaml"))
	})
})

var _ = Describe("Sessions.List", func() {
	It("prints each configured session with its dirs", func() {
		f := newSessionsFixture()

		Expect(f.s.List()).To(Succeed())

		out := f.stdout.String()
		Expect(out).To(ContainSubstring("work"))
		Expect(out).To(ContainSubstring("/p/api"))
		Expect(out).To(ContainSubstring("personal"))
		Expect(out).To(ContainSubstring("/dotfiles"))
	})
})
