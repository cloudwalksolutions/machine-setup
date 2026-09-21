package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
)

var _ = Describe("BackupRoot", func() {
	It("honors the TARS_BACKUP_ROOT override", func() {
		GinkgoT().Setenv("TARS_BACKUP_ROOT", "/mnt/backups")

		Expect(cmd.BackupRoot("/home/u")).To(Equal("/mnt/backups"))
	})

	It("defaults under the user's home, not the repo clone", func() {
		GinkgoT().Setenv("TARS_BACKUP_ROOT", "")

		Expect(cmd.BackupRoot("/home/u")).To(Equal("/home/u/.local/state/tars/backups"))
	})
})
