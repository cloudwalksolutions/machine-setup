package cmd_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/profiles"
)

type spyProfileStore struct {
	file     profiles.File
	err      error
	path     string
	seeded   bool
	appended []profiles.Profile
}

func (s *spyProfileStore) Load() (profiles.File, error) { return s.file, s.err }
func (s *spyProfileStore) Path() string                 { return s.path }
func (s *spyProfileStore) Seed() error {
	s.seeded = true
	return nil
}
func (s *spyProfileStore) Append(p profiles.Profile) error {
	s.appended = append(s.appended, p)
	return nil
}

type profilesFixture struct {
	store  *spyProfileStore
	log    []string
	env    map[string]string
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	p      *cmd.Profiles
}

func newProfilesFixture() *profilesFixture {
	f := &profilesFixture{
		store: &spyProfileStore{
			path: "/cfg/profiles.yaml",
			file: profiles.File{ProjectsDir: "/home/u/Desktop/projects", Profiles: []profiles.Profile{
				{Name: "cloudwalk", Alias: "cws", Email: "me@cloudwalk.example", GitHub: "me-cws"},
				{Name: "oss", Alias: "oss", Email: "me@oss.example", GitHub: "me-oss"},
			}},
		},
		env:    map[string]string{},
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}
	f.p = &cmd.Profiles{
		Home:  "/home/u",
		Store: f.store,
		Apply: func() error {
			f.log = append(f.log, "apply")
			return nil
		},
		SetActive: func(alias string) error {
			f.log = append(f.log, "active:"+alias)
			return nil
		},
		Run: func(name string, args ...string) error {
			f.log = append(f.log, name+" "+strings.Join(args, " "))
			return nil
		},
		LookupEnv: func(key string) (string, bool) {
			v, ok := f.env[key]
			return v, ok
		},
		Stdout: f.stdout,
		Stderr: f.stderr,
	}
	return f
}

var _ = Describe("Profiles.Use", func() {
	It("records the alias as active, re-renders, then switches gh to that account", func() {
		f := newProfilesFixture()

		Expect(f.p.Use("cws")).To(Succeed())

		Expect(f.log).To(Equal([]string{"active:cws", "apply", "gh auth switch --user me-cws"}))
	})

	It("accepts the profile name as well as the alias", func() {
		f := newProfilesFixture()

		Expect(f.p.Use("cloudwalk")).To(Succeed())

		Expect(f.log[0]).To(Equal("active:cws"))
	})

	It("rejects an unknown profile, listing the configured ones, and touches nothing", func() {
		f := newProfilesFixture()

		err := f.p.Use("nope")

		Expect(err).To(MatchError(And(ContainSubstring("nope"), ContainSubstring("cloudwalk (cws)"), ContainSubstring("oss (oss)"))))
		Expect(f.log).To(BeEmpty())
	})

	It("warns that an exported GITHUB_TOKEN overrides whatever gh switches to", func() {
		f := newProfilesFixture()
		f.env["GITHUB_TOKEN"] = "ghp_x"

		Expect(f.p.Use("cws")).To(Succeed())

		Expect(f.stderr.String()).To(ContainSubstring("GITHUB_TOKEN"))
		Expect(f.stderr.String()).To(ContainSubstring("cws.env"))
	})

	It("treats a gh failure as a warning: git identity is already switched", func() {
		f := newProfilesFixture()
		f.p.Run = func(string, ...string) error { return errors.New("gh: not logged in") }

		Expect(f.p.Use("cws")).To(Succeed())

		Expect(f.stderr.String()).To(ContainSubstring("gh: not logged in"))
	})

	It("suggests `tars profiles edit` when there is no config yet", func() {
		f := newProfilesFixture()
		f.store.err = os.ErrNotExist

		err := f.p.Use("cws")

		Expect(err).To(MatchError(And(ContainSubstring("tars profiles edit"), ContainSubstring("/cfg/profiles.yaml"))))
	})
})

var _ = Describe("Profiles.PickAndRun", func() {
	It("offers the aliases and uses the chosen one", func() {
		f := newProfilesFixture()
		picker := &spySessionPicker{choice: "oss"}
		f.p.Picker = picker

		Expect(f.p.PickAndRun()).To(Succeed())

		Expect(picker.options).To(Equal([]string{"cws", "oss"}))
		Expect(f.log[0]).To(Equal("active:oss"))
	})

	It("treats Ctrl+C in the picker as a quiet no-op", func() {
		f := newProfilesFixture()
		f.p.Picker = &spySessionPicker{err: errors.New("user aborted")}

		Expect(f.p.PickAndRun()).To(Succeed())

		Expect(f.log).To(BeEmpty())
	})
})

var _ = Describe("Profiles.Add", func() {
	It("appends the prompted profile to the config and re-renders", func() {
		f := newProfilesFixture()
		newProfile := profiles.Profile{Name: "lab", Alias: "lab", Email: "me@lab.example", GitHub: "me-lab", FullName: "Ada L."}
		f.p.Prompt = func(profiles.Profile) (profiles.Profile, error) { return newProfile, nil }

		Expect(f.p.Add()).To(Succeed())

		Expect(f.store.appended).To(Equal([]profiles.Profile{newProfile}))
		Expect(f.log).To(Equal([]string{"apply"}))
	})

	It("refuses to append an incomplete profile, naming what is missing", func() {
		f := newProfilesFixture()
		f.p.Prompt = func(profiles.Profile) (profiles.Profile, error) {
			return profiles.Profile{Name: "lab", Alias: "lab"}, nil
		}

		err := f.p.Add()

		Expect(err).To(MatchError(ContainSubstring("email")))
		Expect(f.store.appended).To(BeEmpty())
		Expect(f.log).To(BeEmpty())
	})

	It("lets a new profile inherit the shared full_name", func() {
		f := newProfilesFixture()
		f.store.file.FullName = "Ada Lovelace"
		f.p.Prompt = func(profiles.Profile) (profiles.Profile, error) {
			return profiles.Profile{Name: "lab", Alias: "lab", Email: "me@lab.example", GitHub: "me-lab"}, nil
		}

		Expect(f.p.Add()).To(Succeed())

		Expect(f.store.appended).To(HaveLen(1))
	})

	It("stops on a prompt error without touching the config", func() {
		f := newProfilesFixture()
		f.p.Prompt = func(profiles.Profile) (profiles.Profile, error) {
			return profiles.Profile{}, errors.New("user aborted")
		}

		Expect(f.p.Add()).To(Succeed())

		Expect(f.store.appended).To(BeEmpty())
		Expect(f.log).To(BeEmpty())
	})
})

var _ = Describe("Profiles.Edit", func() {
	It("seeds the config, opens it in the editor, then re-renders", func() {
		f := newProfilesFixture()
		var edited string
		f.p.EditFn = func(path string) error {
			edited = path
			f.log = append(f.log, "edit")
			return nil
		}

		Expect(f.p.Edit()).To(Succeed())

		Expect(f.store.seeded).To(BeTrue())
		Expect(edited).To(Equal("/cfg/profiles.yaml"))
		Expect(f.log).To(Equal([]string{"edit", "apply"}))
	})
})

var _ = Describe("Profiles.List", func() {
	It("prints each profile with its dir and email, marking the active one", func() {
		f := newProfilesFixture()
		f.p.ActiveFn = func() (string, error) { return "oss", nil }

		Expect(f.p.List()).To(Succeed())

		Expect(f.stdout.String()).To(Equal(`  cws  cloudwalk  /home/u/Desktop/projects/cloudwalk  me@cloudwalk.example
* oss  oss        /home/u/Desktop/projects/oss        me@oss.example
`))
	})
})

var _ = Describe("Profiles.Show", func() {
	It("prints the named profile's derived paths and identity", func() {
		f := newProfilesFixture()

		Expect(f.p.Show("cws")).To(Succeed())

		Expect(f.stdout.String()).To(Equal(`cloudwalk (cws)
  dir     /home/u/Desktop/projects/cloudwalk
  key     /home/u/.ssh/id_rsa.cws
  email   me@cloudwalk.example
  github  me-cws
`))
	})

	It("defaults to the active profile, and says so when none is active", func() {
		f := newProfilesFixture()
		f.p.ActiveFn = func() (string, error) { return "", nil }

		err := f.p.Show("")

		Expect(err).To(MatchError(ContainSubstring("tars profiles use")))
	})
})

var _ = Describe("NewProfiles", func() {
	It("wires every collaborator for the production command", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		GinkgoT().Setenv("TARS_PROFILES_PATH", filepath.Join(home, "profiles.yaml"))

		p, err := cmd.NewProfiles(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(p.Home).To(Equal(home))
		Expect(p.Store.Path()).To(Equal(filepath.Join(home, "profiles.yaml")))
		Expect(p.Picker).NotTo(BeNil())
		Expect(p.Prompt).NotTo(BeNil())
		Expect(p.EditFn).NotTo(BeNil())
		Expect(p.Apply).NotTo(BeNil())
		Expect(p.SetActive).NotTo(BeNil())
		Expect(p.ActiveFn).NotTo(BeNil())
		Expect(p.Run).NotTo(BeNil())
		Expect(p.LookupEnv).NotTo(BeNil())
	})
})
