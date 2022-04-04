package pipeline_test

import (
	"context"

	"github.com/mokiat/gocrane/internal/pipeline"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {
	var (
		closedCtx context.Context
	)

	BeforeEach(func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		closedCtx = ctx
	})

	Describe("Queue", func() {
		var (
			queue pipeline.Queue[int]
			event int
		)

		BeforeEach(func() {
			queue = make(pipeline.Queue[int])
			event = 132
		})

		Describe("Push", func() {
			It("returns when context is closed", func() {
				Expect(queue.Push(closedCtx, event)).To(BeFalse())

				Consistently(queue).ShouldNot(Receive())
			})

			It("accepts when there is a receiver", func() {
				receivedEvents := make(chan int)
				go func() {
					ev := <-queue
					receivedEvents <- ev
				}()
				Expect(queue.Push(context.Background(), event)).To(BeTrue())

				var receivedEvent int
				Eventually(receivedEvents).Should(Receive(&receivedEvent))
				Expect(receivedEvent).To(Equal(event))
			})
		})

		Describe("Pop", func() {
			It("returns when there is a queued event", func() {
				go func() {
					queue <- event
				}()

				var receivedEvent int
				Expect(queue.Pop(context.Background(), &receivedEvent)).To(BeTrue())
				Expect(receivedEvent).To(Equal(event))
			})

			It("returns when context is closed", func() {
				var receivedEvent int
				Expect(queue.Pop(closedCtx, &receivedEvent)).To(BeFalse())
			})

			It("returns when queue is closed", func() {
				close(queue)
				var receivedEvent int
				Expect(queue.Pop(context.Background(), &receivedEvent)).To(BeFalse())
			})
		})
	})
})
