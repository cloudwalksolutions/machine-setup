package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/forms"
	"tars/internal/pkg"
	"tars/internal/projects"
	"tars/internal/report"
	"tars/internal/sessions"
)

var _ = Describe("IterativeInstaller.InstallAll", func() {
	It("installs only the selected tools, in registry order, continuing past failures", func() {
		var log []string
		stderr := &bytes.Buffer{}
		available := []pkg.Installable{
			&spyInstallable{name: "jq", log: &log},
			&spyInstallable{name: "go", log: &log, err: fmt.Errorf("mirror down")},
			&spyInstallable{name: "fzf", log: &log},
			&spyInstallable{name: "unpicked", log: &log},
		}

		cmd.IterativeInstaller{Report: report.Text{Stdout: &bytes.Buffer{}, Stderr: stderr}}.
			InstallAll(available, []string{"jq", "go", "fzf"})

		Expect(log).To(Equal([]string{"jq", "go", "fzf"}))
		Expect(stderr.String()).To(ContainSubstring("go"))
		Expect(stderr.String()).To(ContainSubstring("mirror down"))
	})
})

var _ = Describe("FileConfigStore", func() {
	It("loads (creating with defaults), saves, and reports its path", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		store := cmd.NewFileConfigStore(path)

		Expect(store.Path()).To(Equal(path))

		cfg, err := store.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(BeARegularFile())

		cfg.Architecture = "riscv"
		Expect(store.Save(cfg)).To(Succeed())

		reloaded, err := store.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(reloaded.Architecture).To(Equal("riscv"))
	})
})

var _ = Describe("Setup.Run failure modes", func() {
	It("fails when the welcome screen errors for a reason other than user-abort", func() {
		f := newFixture()
		f.Welcome.err = fmt.Errorf("tty exploded") // after assemble: Setup holds this spy

		Expect(f.Setup.Run()).To(MatchError(ContainSubstring("welcome")))
	})

	It("fails when the tool picker errors for a reason other than user-abort", func() {
		f := newFixture()
		f.Picker.err = fmt.Errorf("form crashed")

		Expect(f.Setup.Run()).To(MatchError(ContainSubstring("tool picker")))
	})
})

var _ = Describe("composition roots", func() {
	It("NewSetup wires a complete production Setup", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		cfgPath := filepath.Join(GinkgoT().TempDir(), "config.yaml")

		s, err := cmd.NewSetup(report.Text{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}, forms.Headless{}, cfgPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(s.Welcome).NotTo(BeNil())
		Expect(s.Registry).NotTo(BeNil())
		Expect(s.Pull).NotTo(BeNil())
		Expect(s.Wizards).NotTo(BeNil())
		var names []string
		for _, in := range s.Inits {
			names = append(names, in.Name)
			Expect(in.Run).NotTo(BeNil())
		}
		Expect(names).To(Equal([]string{"claude", "pi"}))
	})

	It("NewSessions wires a complete production Sessions", func() {
		s := cmd.NewSessions(&bytes.Buffer{})

		Expect(s.Live).NotTo(BeNil())
		Expect(s.Picker).NotTo(BeNil())
		Expect(s.EditFn).NotTo(BeNil())
		Expect(s.Getwd).NotTo(BeNil())
		Expect(s.Manage).NotTo(BeNil())
	})

	It("NewSessions manages headless under TARS_NO_FORM", func() {
		GinkgoT().Setenv("TARS_NO_FORM", "1")

		Expect(cmd.NewSessions(&bytes.Buffer{}).Headless).To(BeTrue())
	})
})

var _ = Describe("FileProjectCatalog", func() {
	It("discovers the repos under the profiles' projects_dir", func() {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, "api", ".git"), 0o755)).To(Succeed())
		profilesPath := filepath.Join(GinkgoT().TempDir(), "profiles.yaml")
		Expect(os.WriteFile(profilesPath, []byte("projects_dir: "+root+"\n"), 0o644)).To(Succeed())
		GinkgoT().Setenv("TARS_PROFILES_PATH", profilesPath)

		Expect(cmd.NewSessions(&bytes.Buffer{}).Catalog.Projects()).To(Equal([]projects.Project{
			{Name: "api", Dir: filepath.Join(root, "api")},
		}))
	})

	It("seeds, reports, saves and loads the projects file at TARS_PROJECTS_PATH", func() {
		path := filepath.Join(GinkgoT().TempDir(), "projects.yaml")
		GinkgoT().Setenv("TARS_PROJECTS_PATH", path)
		catalog := cmd.NewSessions(&bytes.Buffer{}).Catalog
		saved := projects.File{Projects: map[string]projects.Template{"api": {Windows: []sessions.Window{{Name: "code"}}}}}

		Expect(catalog.Path()).To(Equal(path))
		Expect(catalog.Seed()).To(Succeed())
		Expect(catalog.Templates()).To(Equal(projects.File{}))
		Expect(catalog.SaveTemplates(saved)).To(Succeed())
		Expect(catalog.Templates()).To(Equal(saved))
	})

	It("falls back to the default projects_dir without a profiles file", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		GinkgoT().Setenv("TARS_PROFILES_PATH", filepath.Join(home, "missing.yaml"))
		Expect(os.MkdirAll(filepath.Join(home, "Desktop", "projects", "api", ".git"), 0o755)).To(Succeed())

		Expect(cmd.NewSessions(&bytes.Buffer{}).Catalog.Projects()).To(Equal([]projects.Project{
			{Name: "api", Dir: filepath.Join(home, "Desktop", "projects", "api")},
		}))
	})
})
