package cmd_test

import (
	"bytes"
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/cmd"
	"github.com/cloudwalk/machine-setup/internal/pkg"
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

		cmd.IterativeInstaller{Stdout: &bytes.Buffer{}, Stderr: stderr}.
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

		s, err := cmd.NewSetup(&bytes.Buffer{}, &bytes.Buffer{}, cfgPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(s.Welcome).NotTo(BeNil())
		Expect(s.Registry).NotTo(BeNil())
		Expect(s.Pull).NotTo(BeNil())
	})

	It("NewSessions wires a complete production Sessions", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())

		s, err := cmd.NewSessions(&bytes.Buffer{}, &bytes.Buffer{})

		Expect(err).NotTo(HaveOccurred())
		Expect(s.Store).NotTo(BeNil())
		Expect(s.Opener).NotTo(BeNil())
		Expect(s.Picker).NotTo(BeNil())
		Expect(s.EditFn).NotTo(BeNil())
	})
})

var _ = Describe("FileSessionStore", func() {
	It("seeds, loads, and reports its path", func() {
		path := filepath.Join(GinkgoT().TempDir(), "sessions.yaml")
		GinkgoT().Setenv("MACHINE_SETUP_SESSIONS_PATH", path)
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())

		s, err := cmd.NewSessions(&bytes.Buffer{}, &bytes.Buffer{})
		Expect(err).NotTo(HaveOccurred())

		Expect(s.Store.Path()).To(Equal(path))
		Expect(s.Store.Seed()).To(Succeed())
		f, err := s.Store.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Sessions).NotTo(BeEmpty(), "the seeded example should parse")
	})
})
