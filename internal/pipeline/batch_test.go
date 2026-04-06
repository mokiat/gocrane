package pipeline_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gocrane/internal/pipeline"
)

var _ = Describe("Batch", func() {
	var (
		ctx       context.Context
		ctxCancel func()
		in        pipeline.Queue[pipeline.OldChangeEvent]
		out       pipeline.Queue[pipeline.OldChangeEvent]
		pip       func() error
	)

	BeforeEach(func() {
		ctx, ctxCancel = context.WithCancel(context.Background())
		in = make(pipeline.Queue[pipeline.OldChangeEvent], 1)
		out = make(pipeline.Queue[pipeline.OldChangeEvent], 1)
		pip = pipeline.Batch(ctx, in, out, 200*time.Millisecond)
		go pip()
	})

	AfterEach(func() {
		ctxCancel()
	})

	When("multiple events are pushed in a quick succession", func() {
		BeforeEach(func() {
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"first"}})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"second"}})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"third"}})).To(BeTrue())
		})

		It("produces a combined output event", func() {
			var changeEvent pipeline.OldChangeEvent

			Eventually(out).Should(Receive(&changeEvent))
			Expect(changeEvent.Paths).To(Equal([]string{
				"first", "second", "third",
			}))

			Consistently(out).ShouldNot(Receive(&changeEvent))
		})
	})

	When("events are spread out in time", func() {
		BeforeEach(func() {
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"first"}})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"second"}})).To(BeTrue())
			time.Sleep(500 * time.Millisecond)
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"third"}})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"fourth"}})).To(BeTrue())
		})

		It("produces multiple output event", func() {
			var changeEvent pipeline.OldChangeEvent

			Eventually(out).Should(Receive(&changeEvent))
			Expect(changeEvent.Paths).To(Equal([]string{
				"first", "second",
			}))

			Eventually(out).Should(Receive(&changeEvent))
			Expect(changeEvent.Paths).To(Equal([]string{
				"third", "fourth",
			}))

			Consistently(out).ShouldNot(Receive(&changeEvent))
		})
	})

	When("the pipeline is cancelled", func() {
		BeforeEach(func() {
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"first"}})).To(BeTrue())
			Expect(in.Push(ctx, pipeline.OldChangeEvent{Paths: []string{"second"}})).To(BeTrue())
			ctxCancel()
		})

		It("no longer produces events", func() {
			var changeEvent pipeline.OldChangeEvent
			Consistently(out).ShouldNot(Receive(&changeEvent))
		})
	})
})
