package projects_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/projects"
)

func gitRepo(path string) string {
	Expect(os.MkdirAll(filepath.Join(path, ".git"), 0o755)).To(Succeed())
	return path
}

var _ = Describe("Discover", func() {
	var root string

	BeforeEach(func() { root = GinkgoT().TempDir() })

	It("finds a git repo directly under the projects dir", func() {
		api := gitRepo(filepath.Join(root, "api"))

		Expect(projects.Discover(root)).To(Equal([]projects.Project{{Name: "api", Dir: api}}))
	})

	It("finds git repos one group dir deeper", func() {
		infra := gitRepo(filepath.Join(root, "cloudwalk", "infra"))

		Expect(projects.Discover(root)).To(Equal([]projects.Project{{Name: "infra", Dir: infra}}))
	})

	It("prefixes repos that share a name with their group, sorted by name", func() {
		work := gitRepo(filepath.Join(root, "work", "api"))
		home := gitRepo(filepath.Join(root, "home", "api"))

		Expect(projects.Discover(root)).To(Equal([]projects.Project{
			{Name: "home-api", Dir: home},
			{Name: "work-api", Dir: work},
		}))
	})

	It("names projects tmux-safely", func() {
		site := gitRepo(filepath.Join(root, "example.com"))

		Expect(projects.Discover(root)).To(Equal([]projects.Project{{Name: "example_com", Dir: site}}))
	})

	It("looks no deeper than a group dir", func() {
		gitRepo(filepath.Join(root, "a", "b", "too-deep"))

		Expect(projects.Discover(root)).To(BeEmpty())
	})

	It("does not look inside a repo, and keeps going past it and past files", func() {
		api := gitRepo(filepath.Join(root, "api"))
		gitRepo(filepath.Join(api, "vendor", "lib"))
		Expect(os.WriteFile(filepath.Join(root, "notes.txt"), nil, 0o644)).To(Succeed())
		web := gitRepo(filepath.Join(root, "web"))

		Expect(projects.Discover(root)).To(Equal([]projects.Project{{Name: "api", Dir: api}, {Name: "web", Dir: web}}))
	})

	It("finds nothing when the projects dir does not exist yet", func() {
		Expect(projects.Discover(filepath.Join(root, "missing"))).To(BeEmpty())
	})
})
