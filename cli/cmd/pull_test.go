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

var _ = Describe("ConfirmingPuller.PullAll", func() {
	var (
		applied  []string
		previews *bytes.Buffer
		confirm  *spyConfirmer
		step     func(name, diff string) cmd.PullStep
	)

	BeforeEach(func() {
		applied = nil
		previews = &bytes.Buffer{}
		confirm = &spyConfirmer{answer: true, asked: map[string]string{}}
		step = func(name, diff string) cmd.PullStep {
			return cmd.PullStep{
				Preview: &previewPullable{name: name, out: previews, diff: diff},
				Apply:   &spyPullable{name: name, log: &applied},
			}
		}
	})

	It("shows a changed component's diff and applies it once confirmed", func() {
		puller := cmd.ConfirmingPuller{
			Steps:    []cmd.PullStep{step("zsh", "-old\n+new\n")},
			Previews: previews,
			Confirm:  confirm,
			Report:   report.Text{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}},
		}

		Expect(puller.PullAll()).To(Succeed())

		Expect(confirm.asked).To(Equal(map[string]string{"zsh": "-old\n+new\n"}))
		Expect(applied).To(Equal([]string{"zsh"}))
	})

	It("never asks about a component with nothing to change", func() {
		puller := cmd.ConfirmingPuller{
			Steps:    []cmd.PullStep{step("vim", ""), step("zsh", "+new\n")},
			Previews: previews,
			Confirm:  confirm,
			Report:   report.Text{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}},
		}

		Expect(puller.PullAll()).To(Succeed())

		Expect(confirm.asked).To(HaveKey("zsh"))
		Expect(confirm.asked).NotTo(HaveKey("vim"))
	})

	It("leaves a declined component untouched and says it was skipped", func() {
		confirm.answer = false
		stdout := &bytes.Buffer{}
		puller := cmd.ConfirmingPuller{
			Steps:    []cmd.PullStep{step("zsh", "+new\n")},
			Previews: previews,
			Confirm:  confirm,
			Report:   report.Text{Stdout: stdout, Stderr: &bytes.Buffer{}},
		}

		Expect(puller.PullAll()).To(Succeed())

		Expect(applied).To(BeEmpty())
		Expect(stdout.String()).To(ContainSubstring("skipped zsh"))
	})
})

var _ = Describe("NewConfirmingPuller", func() {
	It("pairs every pullable component with a diff preview of itself", func() {
		opts := components.Options{RepoRoot: GinkgoT().TempDir(), Home: GinkgoT().TempDir(), Stdout: &bytes.Buffer{}}

		p := cmd.NewConfirmingPuller(opts, report.Text{Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})

		Expect(p.Steps).To(HaveLen(len(components.AllPullable(opts))))
		for _, s := range p.Steps {
			Expect(s.Preview.Name()).To(Equal(s.Apply.Name()))
		}
	})
})

var _ = DescribeTable("PullConfirms asks per component only on an interactive, writing pull",
	func(yes, dryRun, headless, tty, want bool) {
		Expect(cmd.PullConfirms(yes, dryRun, headless, tty)).To(Equal(want))
	},
	Entry("a plain pull on a terminal", false, false, false, true, true),
	Entry("--yes", true, false, false, true, false),
	Entry("--dry-run or --diff", false, true, false, true, false),
	Entry("TARS_NO_FORM", false, false, true, true, false),
	Entry("no terminal", false, false, false, false, false),
)

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
