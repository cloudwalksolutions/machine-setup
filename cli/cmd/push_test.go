package cmd_test

import (
	"bytes"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/components"
	"tars/internal/report"
)

type spyPushable struct {
	name string
	log  *[]string
	err  error
}

func (s *spyPushable) Name() string { return s.name }
func (s *spyPushable) Push() error {
	*s.log = append(*s.log, s.name)
	return s.err
}

var _ = Describe("SequentialPusher.PushAll", func() {
	It("pushes each component in order and continues past a failure, reporting it", func() {
		var log []string
		stderr := &bytes.Buffer{}
		pusher := cmd.SequentialPusher{
			Components: []components.Pushable{
				&spyPushable{name: "vim", log: &log},
				&spyPushable{name: "zsh", log: &log, err: errors.New("boom")},
				&spyPushable{name: "terminal", log: &log},
			},
			Report: report.Text{Stdout: &bytes.Buffer{}, Stderr: stderr},
		}

		err := pusher.PushAll()

		Expect(log).To(Equal([]string{"vim", "zsh", "terminal"}))
		Expect(stderr.String()).To(ContainSubstring("zsh"))
		Expect(strings.Count(stderr.String(), "boom")).To(Equal(1))
		Expect(err).To(MatchError(ContainSubstring("zsh")))
	})

	It("returns nil when every component pushes cleanly", func() {
		var log []string
		pusher := cmd.SequentialPusher{
			Components: []components.Pushable{&spyPushable{name: "vim", log: &log}},
			Report:     report.Text{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}},
		}

		Expect(pusher.PushAll()).To(Succeed())
	})
})
