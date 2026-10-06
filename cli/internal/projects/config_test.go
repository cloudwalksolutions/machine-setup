package projects_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/projects"
	"tars/internal/sessions"
)

var _ = Describe("DefaultPath", func() {
	It("honors the TARS_PROJECTS_PATH override", func() {
		GinkgoT().Setenv("TARS_PROJECTS_PATH", "/tmp/custom.yaml")
		Expect(projects.DefaultPath()).To(Equal("/tmp/custom.yaml"))
	})

	It("defaults to ~/.config/tars/projects.yaml", func() {
		GinkgoT().Setenv("TARS_PROJECTS_PATH", "")
		GinkgoT().Setenv("HOME", "/fake/home")
		Expect(projects.DefaultPath()).To(Equal("/fake/home/.config/tars/projects.yaml"))
	})
})

var _ = Describe("Load", func() {
	It("has no templates when the file does not exist yet", func() {
		Expect(projects.Load(filepath.Join(GinkgoT().TempDir(), "projects.yaml"))).To(Equal(projects.File{}))
	})

	It("parses each project's windows", func() {
		path := filepath.Join(GinkgoT().TempDir(), "projects.yaml")
		Expect(os.WriteFile(path, []byte(`projects:
  api:
    windows:
      - {name: code, command: nvim}
      - {name: web, dir: frontend}
`), 0o644)).To(Succeed())

		Expect(projects.Load(path)).To(Equal(projects.File{Projects: map[string]projects.Template{
			"api": {Windows: []sessions.Window{{Name: "code", Command: "nvim"}, {Name: "web", Dir: "frontend"}}},
		}}))
	})
})

var _ = Describe("Save", func() {
	It("writes templates that Load reads back, creating parent dirs", func() {
		path := filepath.Join(GinkgoT().TempDir(), "nested", "projects.yaml")
		f := projects.File{Projects: map[string]projects.Template{
			"api": {Windows: []sessions.Window{{Name: "code", Dir: "src", Command: "nvim"}}},
		}}

		Expect(projects.Save(path, f)).To(Succeed())

		Expect(projects.Load(path)).To(Equal(f))
	})
})

var _ = Describe("Seed", func() {
	It("writes a commented example that loads with no templates", func() {
		path := filepath.Join(GinkgoT().TempDir(), "nested", "projects.yaml")

		Expect(projects.Seed(path)).To(Succeed())

		Expect(os.ReadFile(path)).To(ContainSubstring("# "))
		Expect(projects.Load(path)).To(Equal(projects.File{}))
	})

	It("does not overwrite an existing file", func() {
		path := filepath.Join(GinkgoT().TempDir(), "projects.yaml")
		Expect(os.WriteFile(path, []byte("keep me"), 0o644)).To(Succeed())

		Expect(projects.Seed(path)).To(Succeed())

		Expect(os.ReadFile(path)).To(BeEquivalentTo("keep me"))
	})
})
