package profiles_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/profiles"
)

var cws = profiles.Profile{
	Name: "cloudwalk", Alias: "cws", Email: "me@cloudwalk.example", GitHub: "me-cws", FullName: "Ada Lovelace",
}

var _ = Describe("RenderGitconfig", func() {
	It("renders the identity, the alias-keyed ssh command and the github user", func() {
		Expect(profiles.RenderGitconfig(cws, "/home/u")).To(Equal(`# managed by tars profiles — edit ~/.config/tars/profiles.yaml
[user]
	name = Ada Lovelace
	email = me@cloudwalk.example
[core]
	sshCommand = ssh -i /home/u/.ssh/id_rsa.cws -o IdentitiesOnly=yes
[github]
	user = me-cws
`))
	})
})

var _ = Describe("RenderBlock", func() {
	oss := profiles.Profile{Name: "oss", Alias: "oss", Email: "me@oss.example", GitHub: "me-oss"}
	file := profiles.File{ProjectsDir: "/home/u/Desktop/projects", Profiles: []profiles.Profile{cws, oss}}

	It("scopes each profile's include to its project dir, with the trailing slash git needs", func() {
		Expect(profiles.RenderBlock(file, "", "/home/u/.config/tars/profiles")).To(Equal(`# BEGIN tars profiles
[includeIf "gitdir:/home/u/Desktop/projects/cloudwalk/"]
	path = /home/u/.config/tars/profiles/cws.gitconfig
[includeIf "gitdir:/home/u/Desktop/projects/oss/"]
	path = /home/u/.config/tars/profiles/oss.gitconfig
# END tars profiles
`))
	})

	It("includes the active profile unconditionally first, so dir scopes still win", func() {
		block := profiles.RenderBlock(file, "oss", "/home/u/.config/tars/profiles")

		Expect(block).To(HavePrefix(`# BEGIN tars profiles
[include]
	path = /home/u/.config/tars/profiles/oss.gitconfig
[includeIf "gitdir:/home/u/Desktop/projects/cloudwalk/"]
`))
	})

	It("ignores an active alias that no longer exists", func() {
		Expect(profiles.RenderBlock(file, "gone", "/d")).NotTo(ContainSubstring("[include]\n"))
	})
})

var _ = Describe("RenderEnv", func() {
	It("exports the profile and github user, and leaves a commented slot for the token", func() {
		Expect(profiles.RenderEnv(cws)).To(Equal(`# tars profile "cws": sourced by the shell while active. Personal, never synced.
export TARS_PROFILE="cws"
export GITHUB_USER="me-cws"
# export GITHUB_TOKEN=""
`))
	})
})

var _ = Describe("Splice", func() {
	const block = "# BEGIN tars profiles\n[include]\n\tpath = /p/cws.gitconfig\n# END tars profiles\n"

	It("appends the block after the user's own settings, separated by a blank line", func() {
		out, err := profiles.Splice([]byte("[pull]\n\trebase = false\n"), block)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(Equal("[pull]\n\trebase = false\n\n" + block))
	})

	It("replaces an existing block in place, keeping what surrounds it", func() {
		stale := "# BEGIN tars profiles\n[include]\n\tpath = /p/old.gitconfig\n# END tars profiles\n"
		existing := "[user]\n\tname = x\n\n" + stale + "\n[alias]\n\tco = checkout\n"

		out, err := profiles.Splice([]byte(existing), block)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(Equal("[user]\n\tname = x\n\n" + block + "\n[alias]\n\tco = checkout\n"))
	})

	It("is idempotent: splicing the same block twice yields identical bytes", func() {
		once, err := profiles.Splice([]byte("[user]\n\tname = x\n"), block)
		Expect(err).NotTo(HaveOccurred())

		twice, err := profiles.Splice(once, block)

		Expect(err).NotTo(HaveOccurred())
		Expect(twice).To(Equal(once))
	})

	It("yields just the block when there is no gitconfig yet", func() {
		out, err := profiles.Splice(nil, block)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(Equal(block))
	})

	It("refuses a BEGIN marker with no END rather than guessing where the block stops", func() {
		_, err := profiles.Splice([]byte("# BEGIN tars profiles\n[include]\n"), block)

		Expect(err).To(MatchError(ContainSubstring("END tars profiles")))
	})
})
