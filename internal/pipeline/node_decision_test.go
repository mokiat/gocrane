package pipeline_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/pipeline"
	"github.com/mokiat/gog/filter"
)

var _ = Describe("DecisionNode", func() {
	const (
		eventPath = "/some/file.go"
		otherPath = "/some/other/file.go"
	)

	var (
		ctx       context.Context
		ctxCancel func()

		fakeBuildFilter   filter.Func[string]
		fakeRestartFilter filter.Func[string]

		in  pipeline.Queue[pipeline.ChangeEvent]
		out pipeline.Queue[pipeline.RestartEvent]

		node *pipeline.DecisionNode
	)

	BeforeEach(func() {
		ctx, ctxCancel = context.WithCancel(GinkgoT().Context())

		in = make(pipeline.Queue[pipeline.ChangeEvent], 1)
		out = make(pipeline.Queue[pipeline.RestartEvent], 1)

		fakeBuildFilter = filter.True[string]()
		fakeRestartFilter = filter.True[string]()

		node = pipeline.NewDecisionNode(
			func(path string) bool {
				return fakeBuildFilter(path)
			},
			func(path string) bool {
				return fakeRestartFilter(path)
			},
		)
		go node.Run(ctx, in, out)
	})

	AfterEach(func() {
		ctxCancel()
	})

	When("the pipeline is running", func() {
		JustBeforeEach(func() {
			Expect(in.Push(ctx, pipeline.ChangeEvent{Path: eventPath})).To(BeTrue())
		})

		When("the path matches the rebuild filter", func() {
			BeforeEach(func() {
				fakeBuildFilter = filter.Equal(eventPath)
				fakeRestartFilter = filter.Equal(otherPath)
			})

			It("emits a restart event with ShouldRebuild=true", func() {
				var outEvent pipeline.RestartEvent
				Eventually(out).Should(Receive(&outEvent))
				Expect(outEvent.ShouldRebuild).To(BeTrue())
			})
		})

		When("the path matches only the restart filter", func() {
			BeforeEach(func() {
				fakeBuildFilter = filter.Equal(otherPath)
				fakeRestartFilter = filter.Equal(eventPath)
			})

			It("emits a restart event with ShouldRebuild=false", func() {
				Expect(in.Push(ctx, pipeline.ChangeEvent{Path: eventPath})).To(BeTrue())

				var outEvent pipeline.RestartEvent
				Eventually(out).Should(Receive(&outEvent))
				Expect(outEvent.ShouldRebuild).To(BeFalse())
			})
		})

		When("the path matches neither filter", func() {
			BeforeEach(func() {
				fakeBuildFilter = filter.Equal(otherPath)
				fakeRestartFilter = filter.Equal(otherPath)
			})

			It("does not emit any event", func() {
				Expect(in.Push(ctx, pipeline.ChangeEvent{Path: eventPath})).To(BeTrue())

				var outEvent pipeline.RestartEvent
				Consistently(out).ShouldNot(Receive(&outEvent))
			})
		})

		When("the path matches both filters", func() {
			BeforeEach(func() {
				fakeBuildFilter = filter.Equal(eventPath)
				fakeRestartFilter = filter.Equal(eventPath)
			})

			It("emits a restart event with ShouldRebuild=true (rebuild takes priority)", func() {
				Expect(in.Push(ctx, pipeline.ChangeEvent{Path: eventPath})).To(BeTrue())

				var outEvent pipeline.RestartEvent
				Eventually(out).Should(Receive(&outEvent))
				Expect(outEvent.ShouldRebuild).To(BeTrue())
			})
		})
	})

	When("the pipeline is cancelled", func() {
		BeforeEach(func() {
			ctxCancel()
		})

		It("no longer produces events", func() {
			var outEvent pipeline.RestartEvent
			Consistently(out).ShouldNot(Receive(&outEvent))
		})
	})
})
