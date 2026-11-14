package cmd_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/cmd"
	"tars/internal/components"
	"tars/internal/fsutil"
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
	name    string
	changes []fsutil.Change
	record  func(fsutil.Change)
}

func (p *previewPullable) Name() string { return p.name }
func (p *previewPullable) Pull() error {
	for _, ch := range p.changes {
		p.record(ch)
	}
	return nil
}

type spyReviewer struct {
	got    []components.Plan
	accept []string
	err    error
}

func (s *spyReviewer) Review(plans []components.Plan) ([]string, error) {
	s.got = plans
	return s.accept, s.err
}

var _ = Describe("SequentialPuller.PullAll with a Reviewer", func() {
	var (
		applied  []string
		stdout   *bytes.Buffer
		puller   *cmd.SequentialPuller
		reviewer *spyReviewer
	)

	component := func(name string, changes ...fsutil.Change) {
		puller.Components = append(puller.Components, &spyPullable{name: name, log: &applied})
		puller.Previews = append(puller.Previews, &previewPullable{name: name, changes: changes, record: puller.Record})
	}
	names := func(plans []components.Plan) []string {
		var out []string
		for _, p := range plans {
			out = append(out, p.Component)
		}
		return out
	}

	BeforeEach(func() {
		applied = nil
		stdout = &bytes.Buffer{}
		reviewer = &spyReviewer{}
		puller = &cmd.SequentialPuller{
			Review: reviewer,
			Report: report.Text{Stdout: stdout, Stderr: &bytes.Buffer{}},
		}
	})

	It("reviews every changed component's plan at once, applies the accepted and reports the skipped", func() {
		component("zsh", fsutil.Change{Path: "/h/.zshrc", Kind: fsutil.Overwrite})
		component("nvim", fsutil.Change{Path: "/h/.config/nvim/init.lua", Kind: fsutil.Overwrite})
		reviewer.accept = []string{"zsh"}

		Expect(puller.PullAll()).To(Succeed())

		Expect(names(reviewer.got)).To(Equal([]string{"zsh", "nvim"}))
		Expect(reviewer.got[0].Changes).To(HaveLen(1))
		Expect(applied).To(Equal([]string{"zsh"}))
		Expect(stdout.String()).To(ContainSubstring("skipped nvim"))
	})

	It("never offers an unchanged component, and still pulls it", func() {
		component("vim")
		component("zsh", fsutil.Change{Path: "/h/.zshrc", Kind: fsutil.Create})
		reviewer.accept = []string{"zsh"}

		Expect(puller.PullAll()).To(Succeed())

		Expect(names(reviewer.got)).To(Equal([]string{"zsh"}))
		Expect(applied).To(Equal([]string{"vim", "zsh"}))
	})

	It("skips the review when nothing would change, and still pulls every component", func() {
		component("vim")
		component("zsh")
		reviewer.err = errors.New("review opened with nothing to review")

		Expect(puller.PullAll()).To(Succeed())

		Expect(applied).To(Equal([]string{"vim", "zsh"}))
	})

	It("exposes the plans on their own, for --diff, without pulling anything", func() {
		component("vim")
		component("zsh", fsutil.Change{Path: "/h/.zshrc", Kind: fsutil.Create})

		plans, err := puller.Plans()

		Expect(err).NotTo(HaveOccurred())
		Expect(names(plans)).To(Equal([]string{"zsh"}))
		Expect(applied).To(BeEmpty())
	})

	It("applies nothing when the review is aborted", func() {
		reviewer.err = errors.New("user aborted")
		component("zsh", fsutil.Change{Path: "/h/.zshrc", Kind: fsutil.Create})

		Expect(puller.PullAll()).To(MatchError("user aborted"))

		Expect(applied).To(BeEmpty())
	})
})

var _ = DescribeTable("Interactive reviews only a writing pull on a terminal",
	func(yes, dryRun, headless, ttyIn, ttyOut, want bool) {
		Expect(cmd.Interactive(yes, dryRun, headless, ttyIn, ttyOut)).To(Equal(want))
	},
	Entry("a plain pull on a terminal", false, false, false, true, true, true),
	Entry("--yes", true, false, false, true, true, false),
	Entry("--dry-run", false, true, false, true, true, false),
	Entry("TARS_NO_FORM", false, false, true, true, true, false),
	Entry("stdin is not a terminal", false, false, false, false, true, false),
	Entry("stdout is redirected", false, false, false, true, false, false),
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
