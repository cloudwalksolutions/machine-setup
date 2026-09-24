package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
)

var _ = Describe("Profiles.Pull", func() {
	var (
		tmp        string
		home       string
		configPath string
		opts       components.Options
	)

	newProfiles := func() *components.Profiles {
		c := components.NewProfiles(opts)
		c.ConfigPath = configPath
		return c
	}

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		home = filepath.Join(tmp, "home")
		configPath = filepath.Join(tmp, "profiles.yaml")
		opts = components.Options{
			RepoRoot:   filepath.Join(tmp, "repo"),
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("leaves a machine without a profiles config untouched", func() {
		Expect(newProfiles().Pull()).To(Succeed())

		_, err := os.Stat(home)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	Context("with two profiles configured", func() {
		profilesDir := func() string { return filepath.Join(home, ".config", "tars", "profiles") }

		BeforeEach(func() {
			Expect(os.WriteFile(configPath, []byte(`
full_name: Ada Lovelace
profiles:
  - name: cloudwalk
    alias: cws
    email: me@cloudwalk.example
    github: me-cws
  - name: oss
    alias: oss
    email: me@oss.example
    github: me-oss
`), 0o644)).To(Succeed())
		})

		It("writes one gitconfig per profile, keyed by alias", func() {
			Expect(newProfiles().Pull()).To(Succeed())

			b, err := os.ReadFile(filepath.Join(profilesDir(), "cws.gitconfig"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(ContainSubstring("email = me@cloudwalk.example"))
			Expect(string(b)).To(ContainSubstring("sshCommand = ssh -i " + filepath.Join(home, ".ssh", "id_rsa.cws")))
			Expect(filepath.Join(profilesDir(), "oss.gitconfig")).To(BeARegularFile())
		})

		It("creates ~/.gitconfig holding just the managed block when there is none", func() {
			Expect(newProfiles().Pull()).To(Succeed())

			b, err := os.ReadFile(filepath.Join(home, ".gitconfig"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(`# BEGIN tars profiles
[includeIf "gitdir:` + filepath.Join(home, "Desktop", "projects", "cloudwalk") + `/"]
	path = ` + filepath.Join(profilesDir(), "cws.gitconfig") + `
[includeIf "gitdir:` + filepath.Join(home, "Desktop", "projects", "oss") + `/"]
	path = ` + filepath.Join(profilesDir(), "oss.gitconfig") + `
# END tars profiles
`))
		})

		It("keeps the user's existing gitconfig, archiving it before adding the block", func() {
			Expect(os.MkdirAll(home, 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[pull]\n\trebase = false\n"), 0o644)).To(Succeed())

			Expect(newProfiles().Pull()).To(Succeed())

			b, err := os.ReadFile(filepath.Join(home, ".gitconfig"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(HavePrefix("[pull]\n\trebase = false\n\n# BEGIN tars profiles\n"))
			archived, err := os.ReadFile(filepath.Join(opts.BackupRoot, "profiles", "v1", ".gitconfig"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(archived)).To(Equal("[pull]\n\trebase = false\n"))
		})

		It("is idempotent: a second pull writes nothing and mints no new backup", func() {
			Expect(newProfiles().Pull()).To(Succeed())
			Expect(newProfiles().Pull()).To(Succeed())

			_, err := os.Stat(filepath.Join(opts.BackupRoot, "profiles"))
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("includes the active profile unconditionally once one is set", func() {
			Expect(os.MkdirAll(profilesDir(), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(profilesDir(), "active"), []byte("oss\n"), 0o644)).To(Succeed())

			Expect(newProfiles().Pull()).To(Succeed())

			b, err := os.ReadFile(filepath.Join(home, ".gitconfig"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(ContainSubstring("[include]\n\tpath = " + filepath.Join(profilesDir(), "oss.gitconfig")))
		})

		It("seeds a private env file per profile once and never rewrites it", func() {
			Expect(newProfiles().Pull()).To(Succeed())
			envPath := filepath.Join(profilesDir(), "cws.env")
			info, err := os.Stat(envPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o600)))
			Expect(os.WriteFile(envPath, []byte("export GITHUB_TOKEN=secret\n"), 0o600)).To(Succeed())

			Expect(newProfiles().Pull()).To(Succeed())

			b, err := os.ReadFile(envPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal("export GITHUB_TOKEN=secret\n"))
		})

		It("warns about each missing ssh key instead of failing, since keys arrive later on a fresh machine", func() {
			stderr := &bytes.Buffer{}
			opts.Stderr = stderr
			Expect(os.MkdirAll(filepath.Join(home, ".ssh"), 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(home, ".ssh", "id_rsa.cws"), []byte("KEY"), 0o600)).To(Succeed())

			Expect(newProfiles().Pull()).To(Succeed())

			Expect(stderr.String()).To(ContainSubstring(filepath.Join(home, ".ssh", "id_rsa.oss")))
			Expect(stderr.String()).NotTo(ContainSubstring("id_rsa.cws"))
		})

		It("in dry-run reports every intended file and writes none of them", func() {
			log := &bytes.Buffer{}
			opts.DryRun, opts.Stdout = true, log

			Expect(newProfiles().Pull()).To(Succeed())

			_, err := os.Stat(home)
			Expect(os.IsNotExist(err)).To(BeTrue())
			Expect(log.String()).To(ContainSubstring("would create  " + filepath.Join(profilesDir(), "cws.gitconfig")))
			Expect(log.String()).To(ContainSubstring("would create  " + filepath.Join(profilesDir(), "cws.env")))
			Expect(log.String()).To(ContainSubstring("would create  " + filepath.Join(home, ".gitconfig")))
		})
	})
})
