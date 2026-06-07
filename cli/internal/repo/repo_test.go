package repo_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/repo"
)

var _ = Describe("Find", func() {
	It("honors MACHINE_SETUP_REPO when set", func() {
		tmp := GinkgoT().TempDir()
		GinkgoT().Setenv("MACHINE_SETUP_REPO", tmp)

		got, err := repo.Find()
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(tmp))
	})

	It("walks up from start dir to find a dir containing Makefile + scripts/components", func() {
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "Makefile"), []byte("x"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "scripts", "components"), 0o755)).To(Succeed())

		nested := filepath.Join(root, "a", "b", "c")
		Expect(os.MkdirAll(nested, 0o755)).To(Succeed())

		got, err := repo.FindFrom(nested)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(root))
	})

	It("returns an error when no parent has the markers", func() {
		tmp := GinkgoT().TempDir()
		_, err := repo.FindFrom(tmp)
		Expect(err).To(HaveOccurred())
	})

	It("falls back to walking up from CWD when MACHINE_SETUP_REPO is unset", func() {
		// Build a fake repo with markers + a nested subdir, chdir into it.
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "Makefile"), []byte("x"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "scripts", "components"), 0o755)).To(Succeed())
		nested := filepath.Join(root, "deep", "child")
		Expect(os.MkdirAll(nested, 0o755)).To(Succeed())

		GinkgoT().Setenv("MACHINE_SETUP_REPO", "")
		prev, _ := os.Getwd()
		Expect(os.Chdir(nested)).To(Succeed())
		DeferCleanup(func() { _ = os.Chdir(prev) })

		got, err := repo.Find()
		Expect(err).NotTo(HaveOccurred())
		// macOS tmp paths resolve through /var → /private/var. Compare via EvalSymlinks.
		gotResolved, _ := filepath.EvalSymlinks(got)
		rootResolved, _ := filepath.EvalSymlinks(root)
		Expect(gotResolved).To(Equal(rootResolved))
	})
})
