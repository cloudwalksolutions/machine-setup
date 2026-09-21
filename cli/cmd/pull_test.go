package cmd_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/components"
)

type spyPullable struct {
	name string
	log  *[]string
	err  error
}

func (s *spyPullable) Name() string { return s.name }
func (s *spyPullable) Pull() error {
	*s.log = append(*s.log, s.name)
	return s.err
}

var _ = Describe("SequentialPuller.PullAll", func() {
	It("pulls each component in order, continues past failures, and returns an aggregate error naming them", func() {
		var log []string
		stderr := &bytes.Buffer{}
		puller := cmd.SequentialPuller{
			Components: []components.Component{
				&spyPullable{name: "vim", log: &log},
				&spyPullable{name: "zsh", log: &log, err: errors.New("disk full")},
				&spyPullable{name: "fonts", log: &log, err: errors.New("no sudo")},
			},
			Stdout: &bytes.Buffer{},
			Stderr: stderr,
		}

		err := puller.PullAll()

		Expect(log).To(Equal([]string{"vim", "zsh", "fonts"}))
		Expect(stderr.String()).To(ContainSubstring("disk full"))
		Expect(err).To(MatchError(And(
			ContainSubstring("zsh"),
			ContainSubstring("fonts"),
		)))
	})

	It("returns nil when every component succeeds", func() {
		var log []string
		puller := cmd.SequentialPuller{
			Components: []components.Component{&spyPullable{name: "vim", log: &log}},
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}

		Expect(puller.PullAll()).To(Succeed())
	})
})
