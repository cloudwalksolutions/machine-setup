package sessions_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/sessions"
)

var _ = Describe("Launcher against real byobu", func() {
	var l sessions.Launcher

	BeforeEach(func() {
		if os.Getenv("INTEGRATION") == "" {
			Skip("set INTEGRATION=1 to run byobu integration tests")
		}
		socketDir, err := os.MkdirTemp("/tmp", "tars-tmux") // short: unix socket paths are capped
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, socketDir)
		GinkgoT().Setenv("TMUX_TMPDIR", socketDir)
		GinkgoT().Setenv("TMUX", "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		DeferCleanup(func() { _ = exec.Command("tmux", "kill-server").Run() })
		l = sessions.Launcher{Run: sessions.DefaultRunner(), Output: sessions.DefaultOutput(), LookupEnv: os.LookupEnv}
	})

	It("lists no sessions on a machine where tmux has never run", func() {
		Expect(l.List()).To(BeEmpty())
	})

	It("creates a session from its windows, lists it, then kills it", func() {
		dir := GinkgoT().TempDir()

		Expect(l.Ensure("api", []sessions.Window{{Name: "code", Dir: dir}, {Name: "web", Dir: dir}})).To(Succeed())
		Expect(l.List()).To(Equal([]sessions.Live{{Name: "api", Windows: 2}}))

		Expect(l.Kill("api")).To(Succeed())
		Expect(l.List()).To(BeEmpty())
	})
})
