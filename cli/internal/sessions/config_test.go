package sessions_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/sessions"
)

var _ = Describe("DefaultPath", func() {
	It("honors the MACHINE_SETUP_SESSIONS_PATH override", func() {
		GinkgoT().Setenv("MACHINE_SETUP_SESSIONS_PATH", "/tmp/custom.yaml")
		Expect(sessions.DefaultPath()).To(Equal("/tmp/custom.yaml"))
	})

	It("defaults to ~/.config/.machine-setup/sessions.yaml", func() {
		GinkgoT().Setenv("MACHINE_SETUP_SESSIONS_PATH", "")
		GinkgoT().Setenv("HOME", "/fake/home")
		Expect(sessions.DefaultPath()).To(Equal("/fake/home/.config/.machine-setup/sessions.yaml"))
	})
})

var _ = Describe("Seed", func() {
	It("writes a commented example config, creating parent dirs", func() {
		path := filepath.Join(GinkgoT().TempDir(), "nested", "sessions.yaml")

		Expect(sessions.Seed(path)).To(Succeed())

		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("sessions:"))
		Expect(string(data)).To(ContainSubstring("#"))
	})

	It("does not overwrite an existing file", func() {
		path := filepath.Join(GinkgoT().TempDir(), "sessions.yaml")
		Expect(os.WriteFile(path, []byte("keep me"), 0o644)).To(Succeed())

		Expect(sessions.Seed(path)).To(Succeed())

		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("keep me"))
	})
})

var _ = Describe("WindowName", func() {
	It("uses the dir basename with '.' and ':' sanitized to '_'", func() {
		Expect(sessions.WindowName("/projects/machine-setup")).To(Equal("machine-setup"))
		Expect(sessions.WindowName("/projects/api.v2:beta")).To(Equal("api_v2_beta"))
	})
})

var _ = Describe("Load", func() {
	var path string

	BeforeEach(func() {
		path = filepath.Join(GinkgoT().TempDir(), "sessions.yaml")
	})

	write := func(content string) {
		Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())
	}

	It("expands ~ in dirs to the user's home", func() {
		GinkgoT().Setenv("HOME", "/fake/home")
		write(`sessions:
  - name: work
    dirs:
      - ~/projects/api
      - "~"
`)

		f, err := sessions.Load(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Sessions[0].Dirs).To(Equal([]string{"/fake/home/projects/api", "/fake/home"}))
	})

	It("rejects session names containing '.' or ':' (tmux target syntax)", func() {
		write(`sessions:
  - name: bad.name
    dirs: ["/x"]
`)

		_, err := sessions.Load(path)
		Expect(err).To(MatchError(ContainSubstring("bad.name")))
	})

	It("rejects sessions without dirs", func() {
		write(`sessions:
  - name: empty
`)

		_, err := sessions.Load(path)
		Expect(err).To(MatchError(ContainSubstring("empty")))
	})

	It("parses session names and dirs from yaml", func() {
		write(`sessions:
  - name: cloudwalk
    dirs:
      - /projects/machine-setup
      - /projects/api
  - name: personal
    dirs:
      - /dotfiles
`)

		f, err := sessions.Load(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Sessions).To(Equal([]sessions.Session{
			{Name: "cloudwalk", Dirs: []string{"/projects/machine-setup", "/projects/api"}},
			{Name: "personal", Dirs: []string{"/dotfiles"}},
		}))
	})
})
