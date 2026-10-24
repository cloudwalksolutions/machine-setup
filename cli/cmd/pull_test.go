package cmd_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/components"
	"tars/internal/report"
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

type previewPullable struct {
	name string
	out  *bytes.Buffer
	diff string
}

func (p *previewPullable) Name() string { return p.name }
func (p *previewPullable) Pull() error {
	p.out.WriteString(p.diff)
	return nil
}

type spyConfirmer struct {
	answer bool
	asked  map[string]string
}

func (s *spyConfirmer) ConfirmApply(name, diff string) (bool, error) {
	s.asked[name] = diff
	return s.answer, nil
}

var _ = Describe("SequentialPuller.PullAll with Confirm", func() {
	var (
		applied []string
		stdout  *bytes.Buffer
		puller  cmd.SequentialPuller
		confirm *spyConfirmer
	)

	component := func(name, diff string) {
		puller.Components = append(puller.Components, &spyPullable{name: name, log: &applied})
		puller.Previews = append(puller.Previews, &previewPullable{name: name, out: puller.Diffs, diff: diff})
	}

	BeforeEach(func() {
		applied = nil
		stdout = &bytes.Buffer{}
		confirm = &spyConfirmer{answer: true, asked: map[string]string{}}
		puller = cmd.SequentialPuller{
			Diffs:   &bytes.Buffer{},
			Confirm: confirm,
			Report:  report.Text{Stdout: stdout, Stderr: &bytes.Buffer{}},
		}
	})

	It("shows a changed component's diff and applies it once confirmed", func() {
		component("zsh", "-old\n+new\n")

		Expect(puller.PullAll()).To(Succeed())

		Expect(confirm.asked).To(Equal(map[string]string{"zsh": "-old\n+new\n"}))
		Expect(applied).To(Equal([]string{"zsh"}))
	})

	It("never asks about a component with nothing to change", func() {
		component("vim", "")
		component("zsh", "+new\n")

		Expect(puller.PullAll()).To(Succeed())

		Expect(confirm.asked).To(HaveKey("zsh"))
		Expect(confirm.asked).NotTo(HaveKey("vim"))
	})

	It("leaves a declined component untouched and says it was skipped", func() {
		confirm.answer = false
		component("zsh", "+new\n")

		Expect(puller.PullAll()).To(Succeed())

		Expect(applied).To(BeEmpty())
		Expect(stdout.String()).To(ContainSubstring("skipped zsh"))
	})
})

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
			Report: report.Text{Stdout: &bytes.Buffer{}, Stderr: stderr},
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
			Report:     report.Text{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}},
		}

		Expect(puller.PullAll()).To(Succeed())
	})
})
