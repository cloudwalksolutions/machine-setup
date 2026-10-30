package cmd_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
)

var _ = Describe("ResolveRepo", func() {
	It("returns the clone when one is found", func() {
		GinkgoT().Setenv("TARS_REPO", "/some/clone")

		root, err := cmd.ResolveRepo(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		Expect(root).To(Equal("/some/clone"))
	})

	It("materializes the embedded dotfiles when no clone is found", func() {
		GinkgoT().Setenv("TARS_REPO", "")
		home := GinkgoT().TempDir()

		// Run from a dir with no repo markers above it so the walk fails.
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(GinkgoT().TempDir())).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(cwd) })

		root, err := cmd.ResolveRepo(home)
		Expect(err).NotTo(HaveOccurred())
		Expect(root).To(Equal(filepath.Join(home, ".local", "share", "tars", "repo")))
		Expect(filepath.Join(root, "zsh", "zshrc")).To(BeARegularFile())
		Expect(filepath.Join(root, "nvim", "init.lua")).To(BeARegularFile())
	})
})
