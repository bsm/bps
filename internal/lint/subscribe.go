package lint

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bsm/bps"
	"github.com/onsi/ginkgo"
	Ω "github.com/onsi/gomega"
)

// SubscriberInput for the shared test.
type SubscriberInput struct {
	// Subject should return a subject to be tested, with given topic/messages seeded.
	// Caller should handle all the subject teardown/cleanup.
	Subject func(topic string, messages []bps.SubMessage) bps.Subscriber
}

// Subscriber lints subscribers.
func Subscriber(input *SubscriberInput) {
	var subject bps.Subscriber
	var handler *mockHandler
	var ctx = context.Background()

	var (
		topic        string
		unknownTopic string
	)

	ginkgo.BeforeEach(func() {
		cycle := time.Now().UnixNano()
		topic = fmt.Sprintf("bps-unittest-topic-%d-a", cycle)
		unknownTopic = fmt.Sprintf("bps-unittest-topic-%d-unknown", cycle)

		subject = input.Subject(topic, []bps.SubMessage{
			bps.RawSubMessage("message-1"),
			bps.RawSubMessage("message-2"),
		})
		handler = &mockHandler{}
	})

	ginkgo.It("should subscribe", func() {
		Ω.Expect(subject.Subscribe(ctx, topic, handler)).To(Ω.Succeed())
		Ω.Expect(extractData(handler.Received)).To(Ω.Equal([][]byte{
			[]byte("message-1"),
			[]byte("message-2"),
		}))
	})

	ginkgo.It("should stop on bps.Done", func() {
		// allow error wrapping:
		handler.Err = fmt.Errorf("wrapped %w", bps.Done)

		Ω.Expect(subject.Subscribe(ctx, topic, handler)).To(Ω.Succeed())

		// only first message handled - before error is returned:
		Ω.Expect(extractData(handler.Received)).To(Ω.Equal([][]byte{
			[]byte("message-1"),
		}))
	})

	ginkgo.It("should return handler error", func() {
		expectedErr := errors.New("foo")
		handler.Err = expectedErr

		actualErr := subject.Subscribe(ctx, topic, handler)

		// allow error wrapping:
		Ω.Expect(errors.Is(actualErr, expectedErr)).To(Ω.BeTrue())

		// only first message handled - before error is returned:
		Ω.Expect(extractData(handler.Received)).To(Ω.Equal([][]byte{
			[]byte("message-1"),
		}))
	})

	ginkgo.It("should stop when context is cancelled", func() {
		cancellableCtx, cancel := context.WithCancel(ctx)
		handler.AfterHandle = cancel

		err := subject.Subscribe(cancellableCtx, topic, handler)

		// allow error wrapping:
		Ω.Expect(errors.Is(err, context.Canceled)).To(Ω.BeTrue())

		// only first message handled - before context is cancelled:
		Ω.Expect(extractData(handler.Received)).To(Ω.Equal([][]byte{
			[]byte("message-1"),
		}))
	})

	ginkgo.It("should do nothing for unknown topicss", func() {
		Ω.Expect(subject.Subscribe(ctx, unknownTopic, handler)).To(Ω.Succeed())
		Ω.Expect(handler.Received).To(Ω.BeEmpty())
	})
}

type mockHandler struct {
	Err         error  // error to return after handling (first) message
	AfterHandle func() // after message handled callback
	Received    []bps.SubMessage
}

func (h *mockHandler) Handle(msg bps.SubMessage) error {
	h.Received = append(h.Received, msg)
	if cb := h.AfterHandle; cb != nil {
		cb()
	}
	return h.Err
}

func extractData(msgs []bps.SubMessage) [][]byte {
	data := make([][]byte, 0, len(msgs))
	for _, msg := range msgs {
		data = append(data, msg.Data())
	}
	return data
}
