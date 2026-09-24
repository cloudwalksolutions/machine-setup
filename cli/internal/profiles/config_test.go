package profiles_test

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/profiles"
)

var _ = Describe("DefaultPath", func() {
	It("honors the TARS_PROFILES_PATH override", func() {
		GinkgoT().Setenv("TARS_PROFILES_PATH", "/tmp/custom.yaml")
		Expect(profiles.DefaultPath()).To(Equal("/tmp/custom.yaml"))
	})

	It("defaults to ~/.config/tars/profiles.yaml", func() {
		GinkgoT().Setenv("TARS_PROFILES_PATH", "")
		GinkgoT().Setenv("HOME", "/fake/home")
		Expect(profiles.DefaultPath()).To(Equal("/fake/home/.config/tars/profiles.yaml"))
	})
})

const oneProfile = `
profiles:
  - name: cloudwalk
    alias: cws
    email: me@cloudwalk.example
    github: me-cws
`

var _ = Describe("Load", func() {
	var (
		home string
		path string
	)

	writeConfig := func(content string) {
		Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())
	}

	load := func() profiles.File {
		f, err := profiles.Load(path, home)
		Expect(err).NotTo(HaveOccurred())
		return f
	}

	BeforeEach(func() {
		tmp := GinkgoT().TempDir()
		home = filepath.Join(tmp, "home")
		path = filepath.Join(tmp, "profiles.yaml")
	})

	It("parses name, alias, email and github for each profile", func() {
		writeConfig(oneProfile)

		p := load().Profiles[0]

		Expect(p.Name).To(Equal("cloudwalk"))
		Expect(p.Alias).To(Equal("cws"))
		Expect(p.Email).To(Equal("me@cloudwalk.example"))
		Expect(p.GitHub).To(Equal("me-cws"))
	})

	It("places each profile's dir under ~/Desktop/projects/<name> by default", func() {
		writeConfig(oneProfile)

		f := load()

		Expect(f.ProjectsDir).To(Equal(filepath.Join(home, "Desktop", "projects")))
		Expect(f.Profiles[0].Dir(f.ProjectsDir)).To(Equal(filepath.Join(home, "Desktop", "projects", "cloudwalk")))
	})

	It("expands ~ in a configured projects_dir", func() {
		writeConfig("projects_dir: ~/src\n" + oneProfile)

		Expect(load().ProjectsDir).To(Equal(filepath.Join(home, "src")))
	})

	It("derives the ssh key path from the alias: ~/.ssh/id_rsa.<alias>", func() {
		writeConfig(oneProfile)

		Expect(load().Profiles[0].KeyPath(home)).To(Equal(filepath.Join(home, ".ssh", "id_rsa.cws")))
	})

	It("gives every profile the top-level full_name unless it sets its own", func() {
		writeConfig(`
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
    full_name: Ada L.
`)

		f := load()

		Expect(f.Profiles[0].FullName).To(Equal("Ada Lovelace"))
		Expect(f.Profiles[1].FullName).To(Equal("Ada L."))
	})

	It("sorts profiles by alias so rendered output is stable", func() {
		writeConfig(`
profiles:
  - name: zeta
    alias: zz
    email: z@example.com
    github: z
  - name: cloudwalk
    alias: cws
    email: me@cloudwalk.example
    github: me-cws
`)

		f := load()

		Expect(f.Profiles[0].Alias).To(Equal("cws"))
		Expect(f.Profiles[1].Alias).To(Equal("zz"))
	})

	DescribeTable("rejects a name or alias that is not a safe lowercase identifier",
		func(name, alias, offending string) {
			writeConfig(`
profiles:
  - name: "` + name + `"
    alias: "` + alias + `"
    email: me@cloudwalk.example
    github: me-cws
`)

			_, err := profiles.Load(path, home)

			Expect(err).To(MatchError(ContainSubstring(offending)))
		},
		Entry("uppercase name", "CloudWalk", "cws", "CloudWalk"),
		Entry("space in alias", "cloudwalk", "c ws", "c ws"),
		Entry("dot in name", "cloud.walk", "cws", "cloud.walk"),
		Entry("slash in alias", "cloudwalk", "c/ws", "c/ws"),
	)

	DescribeTable("rejects a profile missing a required field",
		func(body, missing string) {
			writeConfig("profiles:\n  - name: cloudwalk\n    alias: cws\n" + body)

			_, err := profiles.Load(path, home)

			Expect(err).To(MatchError(And(ContainSubstring("cws"), ContainSubstring(missing))))
		},
		Entry("no email", "    github: me-cws\n", "email"),
		Entry("no github", "    email: me@cloudwalk.example\n", "github"),
	)

	DescribeTable("rejects two profiles sharing an identifier",
		func(secondName, secondAlias, duplicate string) {
			writeConfig(oneProfile + `
  - name: ` + secondName + `
    alias: ` + secondAlias + `
    email: other@example.com
    github: other
`)

			_, err := profiles.Load(path, home)

			Expect(err).To(MatchError(ContainSubstring(duplicate)))
		},
		Entry("same name", "cloudwalk", "cw2", "cloudwalk"),
		Entry("same alias", "cloudwalk2", "cws", "cws"),
	)

	It("finds a profile by alias or by name, and reports a miss", func() {
		writeConfig(oneProfile)
		f := load()

		byAlias, found := f.Find("cws")
		Expect(found).To(BeTrue())
		Expect(byAlias.Name).To(Equal("cloudwalk"))

		byName, found := f.Find("cloudwalk")
		Expect(found).To(BeTrue())
		Expect(byName.Alias).To(Equal("cws"))

		_, found = f.Find("nope")
		Expect(found).To(BeFalse())
	})

	It("reports os.ErrNotExist when there is no config yet", func() {
		_, err := profiles.Load(path, home)

		Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
	})
})

var _ = Describe("Seed", func() {
	It("writes a commented example that Load accepts, creating parent dirs", func() {
		tmp := GinkgoT().TempDir()
		path := filepath.Join(tmp, "nested", "profiles.yaml")

		Expect(profiles.Seed(path)).To(Succeed())

		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("#"))
		f, err := profiles.Load(path, filepath.Join(tmp, "home"))
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Profiles).NotTo(BeEmpty())
	})

	It("does not overwrite an existing file", func() {
		path := filepath.Join(GinkgoT().TempDir(), "profiles.yaml")
		Expect(os.WriteFile(path, []byte("keep me"), 0o644)).To(Succeed())

		Expect(profiles.Seed(path)).To(Succeed())

		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("keep me"))
	})
})

var _ = Describe("Append", func() {
	It("adds a profile to an existing config that Load reads back, keeping the user's comments", func() {
		tmp := GinkgoT().TempDir()
		path := filepath.Join(tmp, "profiles.yaml")
		Expect(os.WriteFile(path, []byte("# mine\n"+oneProfile), 0o644)).To(Succeed())

		Expect(profiles.Append(path, profiles.Profile{
			Name: "oss", Alias: "oss", Email: "me@oss.example", GitHub: "me-oss",
		})).To(Succeed())

		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(HavePrefix("# mine\n"))
		f, err := profiles.Load(path, filepath.Join(tmp, "home"))
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Profiles).To(HaveLen(2))
		Expect(f.Profiles[1].Alias).To(Equal("oss"))
	})
})
