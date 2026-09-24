package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/forms"
)

var _ = Describe("ProjectInit.Run", func() {
	var (
		tmp     string
		dir     string
		asker   *spyProjectAsker
		project *cmd.ProjectInit
	)

	read := func(path string) string {
		b, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred(), path)
		return string(b)
	}

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		dir = filepath.Join(tmp, "my-app")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		tpl := filepath.Join(tmp, "AGENTS.project.md")
		Expect(os.WriteFile(tpl, []byte("# {{.Name}}\n\n{{.Description}}\n\n```\n{{.TestCommand}}\n```\n"), 0o644)).To(Succeed())
		asker = &spyProjectAsker{answer: forms.ProjectAnswers{
			Agents: []string{"claude", "gemini"}, Description: "Does things", TestCommand: "make test",
		}}
		project = &cmd.ProjectInit{
			Asker:      asker,
			Template:   tpl,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
		}
	})

	It("renders AGENTS.md from the template and a pointer file per selected agent", func() {
		Expect(project.Run(dir)).To(Succeed())

		Expect(read(filepath.Join(dir, "AGENTS.md"))).To(Equal("# my-app\n\nDoes things\n\n```\nmake test\n```\n"))
		Expect(read(filepath.Join(dir, "CLAUDE.md"))).To(Equal("@AGENTS.md\n"))
		Expect(read(filepath.Join(dir, "GEMINI.md"))).To(Equal("@./AGENTS.md\n"))
	})

	It("refuses to overwrite an existing AGENTS.md unless forced", func() {
		Expect(os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("KEEP"), 0o644)).To(Succeed())

		Expect(project.Run(dir)).To(MatchError(ContainSubstring("--force")))

		Expect(read(filepath.Join(dir, "AGENTS.md"))).To(Equal("KEEP"))
		Expect(filepath.Join(dir, "CLAUDE.md")).NotTo(BeAnExistingFile())
	})

	It("fails when the template is missing and writes nothing", func() {
		project.Template = filepath.Join(tmp, "nope.md")

		Expect(project.Run(dir)).NotTo(Succeed())
		Expect(filepath.Join(dir, "AGENTS.md")).NotTo(BeAnExistingFile())
		Expect(filepath.Join(dir, "CLAUDE.md")).NotTo(BeAnExistingFile())
	})

	It("backs up the existing AGENTS.md under project before overwriting when forced", func() {
		Expect(os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("OLD"), 0o644)).To(Succeed())
		project.Force = true

		Expect(project.Run(dir)).To(Succeed())

		Expect(read(filepath.Join(project.BackupRoot, "project", "v1", "AGENTS.md"))).To(Equal("OLD"))
		Expect(read(filepath.Join(dir, "AGENTS.md"))).To(HavePrefix("# my-app"))
	})
})
