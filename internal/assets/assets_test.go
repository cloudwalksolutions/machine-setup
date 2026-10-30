package assets_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/assets"
)

var _ = Describe("Materialize", func() {
	var dst string

	BeforeEach(func() {
		dst = filepath.Join(GinkgoT().TempDir(), "repo")
	})

	It("extracts the full tree with a version marker, marking bin scripts executable", func() {
		Expect(assets.Materialize(dst, "v1.2.3")).To(Succeed())

		Expect(filepath.Join(dst, "zsh", "zshrc")).To(BeARegularFile())
		Expect(filepath.Join(dst, "byobu", ".tmux.conf")).To(BeARegularFile())
		Expect(filepath.Join(dst, "nvim", "init.lua")).To(BeARegularFile())
		Expect(filepath.Join(dst, "agents", "edit-guard.json")).To(BeARegularFile())
		Expect(filepath.Join(dst, "monokai.lua")).To(BeARegularFile())

		info, err := os.Stat(filepath.Join(dst, "byobu", "bin", "1_git"))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()&0o111).NotTo(BeZero(), "bin scripts must be executable")

		info, err = os.Stat(filepath.Join(dst, "claude", "hooks", "block-unreviewable-edits.sh"))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()&0o111).NotTo(BeZero(), "hook scripts must be executable")

		marker, err := os.ReadFile(filepath.Join(dst, ".tars-version"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(marker)).To(Equal("v1.2.3"))
	})

	It("skips re-extraction when the version marker already matches", func() {
		Expect(assets.Materialize(dst, "v1.2.3")).To(Succeed())
		sentinel := filepath.Join(dst, "zsh", "sentinel")
		Expect(os.WriteFile(sentinel, []byte("keep"), 0o644)).To(Succeed())

		Expect(assets.Materialize(dst, "v1.2.3")).To(Succeed())

		Expect(sentinel).To(BeARegularFile())
	})

	It("re-extracts for dev builds even when the marker matches", func() {
		Expect(assets.Materialize(dst, "dev")).To(Succeed())
		mutated := filepath.Join(dst, "zsh", "zshrc")
		Expect(os.WriteFile(mutated, []byte("mutated"), 0o644)).To(Succeed())

		Expect(assets.Materialize(dst, "dev")).To(Succeed())

		b, err := os.ReadFile(mutated)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).NotTo(Equal("mutated"))
	})
})
