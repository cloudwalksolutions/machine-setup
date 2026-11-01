package components_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
	"tars/internal/fsutil"
)

var _ = Describe("Plan", func() {
	It("summarizes a component's changes by kind, in create, modify, remove, set order", func() {
		plan := components.Plan{Component: "nvim", Changes: []fsutil.Change{
			{Kind: fsutil.Overwrite}, {Kind: fsutil.Remove}, {Kind: fsutil.Overwrite}, {Kind: fsutil.Create},
		}}

		Expect(plan.Summary()).To(Equal("1 create, 2 modify, 1 remove"))
		Expect(components.Plan{Component: "terminal", Changes: []fsutil.Change{{Kind: fsutil.Set}}}.Summary()).To(Equal("1 setting"))
	})

	It("totals the plans as a Plan line to preview and a recap once applied", func() {
		plans := []components.Plan{
			{Component: "zsh", Changes: []fsutil.Change{{Kind: fsutil.Overwrite}, {Kind: fsutil.Create}}},
			{Component: "nvim", Changes: []fsutil.Change{{Kind: fsutil.Overwrite}, {Kind: fsutil.Remove}}},
		}

		Expect(components.PlanLine(plans)).To(Equal("Plan: 1 to create, 2 to modify, 1 to remove · 2 components"))
		Expect(components.Recap(plans, 7)).To(Equal("Pull complete: 1 created, 2 overwritten, 1 removed, 7 components unchanged."))
		Expect(components.Recap(nil, 9)).To(Equal("Pull complete: nothing to change, 9 components unchanged."))
	})
})
