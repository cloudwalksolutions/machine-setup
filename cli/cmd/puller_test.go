package cmd_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/cmd"
	"github.com/cloudwalk/machine-setup/internal/components"
)

type failingComponent struct {
	name string
	err  error
}

func (c *failingComponent) Name() string { return c.name }
func (c *failingComponent) Pull() error  { return c.err }

var _ = Describe("SequentialPuller.PullAll", func() {
	var stdout, stderr *bytes.Buffer

	newPuller := func(comps ...components.Component) cmd.SequentialPuller {
		return cmd.SequentialPuller{Components: comps, Stdout: stdout, Stderr: stderr}
	}

	BeforeEach(func() {
		stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
	})

	It("returns nil when every component succeeds", func() {
		p := newPuller(
			&failingComponent{name: "vim", err: nil},
			&failingComponent{name: "zsh", err: nil},
		)
		Expect(p.PullAll()).To(Succeed())
	})

	It("returns an error naming the failed component", func() {
		p := newPuller(&failingComponent{name: "zsh", err: errors.New("boom")})

		err := p.PullAll()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("zsh"))
	})

	It("keeps pulling after a failure and reports every failure", func() {
		p := newPuller(
			&failingComponent{name: "vim", err: errors.New("bad vim")},
			&failingComponent{name: "zsh", err: nil},
			&failingComponent{name: "nvim", err: errors.New("bad nvim")},
		)

		err := p.PullAll()

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("vim"))
		Expect(err.Error()).To(ContainSubstring("nvim"))
		Expect(stdout.String()).To(ContainSubstring("zsh"))
	})
})
