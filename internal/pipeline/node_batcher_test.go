package pipeline_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/pipeline"
)

var _ = Describe("BatcherNode", func() {
	var (
		ctx       context.Context
		ctxCancel func()

		in  pipeline.Queue[pipeline.RestartEvent]
		out pipeline.Queue[pipeline.RestartEvent]

		node *pipeline.BatcherNode
	)

	BeforeEach(func() {
		ctx, ctxCancel = context.WithCancel(GinkgoT().Context())

		in = make(pipeline.Queue[pipeline.RestartEvent], 1)
		out = make(pipeline.Queue[pipeline.RestartEvent], 1)

		node = pipeline.NewBatcherNode(200 * time.Millisecond)
		go node.Run(ctx, in, out)
	})

	AfterEach(func() {
		ctxCancel()
	})

	When("multiple events are pushed in a quick succession", func() {
		BeforeEach(func() {
			Expect(in.Push(ctx, pipeline.RestartEvent{false})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.RestartEvent{true})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.RestartEvent{false})).To(BeTrue())
		})

		It("produces a combined output event", func() {
			var outputEvent pipeline.RestartEvent

			Eventually(out).Should(Receive(&outputEvent))
			Expect(outputEvent.ShouldRebuild).To(BeTrue())

			Consistently(out).ShouldNot(Receive(&outputEvent))
		})
	})

	When("events are spread out in time", func() {
		BeforeEach(func() {
			Expect(in.Push(ctx, pipeline.RestartEvent{true})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.RestartEvent{false})).To(BeTrue())
			time.Sleep(500 * time.Millisecond)
			Expect(in.Push(ctx, pipeline.RestartEvent{false})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.RestartEvent{false})).To(BeTrue())
		})

		It("produces multiple output event", func() {
			var outputEvent pipeline.RestartEvent

			Eventually(out).Should(Receive(&outputEvent))
			Expect(outputEvent.ShouldRebuild).To(BeTrue())

			Eventually(out).Should(Receive(&outputEvent))
			Expect(outputEvent.ShouldRebuild).To(BeFalse())

			Consistently(out).ShouldNot(Receive(&outputEvent))
		})
	})

	When("the pipeline is cancelled", func() {
		BeforeEach(func() {
			Expect(in.Push(ctx, pipeline.RestartEvent{false})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.RestartEvent{true})).To(BeTrue())
			ctxCancel()
		})

		It("no longer produces events", func() {
			var outputEvent pipeline.RestartEvent
			Consistently(out).ShouldNot(Receive(&outputEvent))
		})
	})
})
