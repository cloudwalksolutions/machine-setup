package assets_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/assets"
	"tars/internal/repo"
)

// The embedded tree must stay in sync with the repo's dotfile dirs. When this
// fails, run `make sync-assets` and commit the refreshed mirror.
var _ = Describe("embedded tree drift guard", func() {
	It("includes the repo's root-level files (monokai.lua) byte-for-byte", func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		root, err := repo.FindFrom(cwd)
		Expect(err).NotTo(HaveOccurred())

		for _, name := range assets.RootFiles {
			want, err := os.ReadFile(filepath.Join(root, name))
			Expect(err).NotTo(HaveOccurred())

			embedded, err := assets.Files(name)
			Expect(err).NotTo(HaveOccurred())
			Expect(embedded[name]).To(Equal(want),
				"%s drifted from the repo — run `make sync-assets`", name)
		}
	})

	It("matches the repo's dotfile dirs byte-for-byte, both directions", func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		root, err := repo.FindFrom(cwd)
		Expect(err).NotTo(HaveOccurred())

		for _, dir := range assets.Dirs {
			repoFiles := map[string][]byte{}
			repoDir := filepath.Join(root, dir)
			Expect(filepath.WalkDir(repoDir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				repoFiles[rel] = b
				return nil
			})).To(Succeed())

			embedded, err := assets.Files(dir)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(embedded)).To(Equal(len(repoFiles)),
				"%s: embedded file count differs from repo — run `make sync-assets`", dir)
			for rel, want := range repoFiles {
				Expect(embedded[rel]).To(Equal(want),
					"%s drifted from the repo — run `make sync-assets`", rel)
			}
		}
	})
})

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

		info, err := os.Stat(filepath.Join(dst, "byobu", "bin", "1_git"))
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()&0o111).NotTo(BeZero(), "bin scripts must be executable")

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
