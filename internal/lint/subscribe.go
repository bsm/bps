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
	// Subject holds a subscriber to be tested.
	Subject bps.Subscriber
	// SeededMessages contain all the messages, seeded for this subscriber.
	// Must contain at 2+ messages.
	// Seed/consume order must be preserved.
	SeededMessages []bps.SubMessage
	// Setup is called before testing the subscriber.
	Setup func(topic string) error
	// Teardown is called after testing the subscriber.
	Teardown func(topic string) error
}

// Subcriber lints subscribers.
func Subcriber(input *SubscriberInput) {
	var subject bps.Subscriber
	var handler *mockHandler
	var ctx = context.Background()

	var (
		knownTopic   string
		unknownTopic string
	)

	ginkgo.BeforeEach(func() {
		// sanity check - require at least 2 messages seeded for subscriber for proper stop/abort testing:
		Ω.Expect(input.SeededMessages).To(Ω.HaveLen(2))

		subject = input.Subject
		handler = &mockHandler{}

		cycle := time.Now().UnixNano()
		knownTopic = fmt.Sprintf("bps-unittest-topic-%d-a", cycle)
		unknownTopic = fmt.Sprintf("bps-unittest-topic-%d-unknown", cycle)

		if input.Setup != nil {
			Ω.Expect(input.Setup(knownTopic)).To(Ω.Succeed())
		}
	})

	ginkgo.AfterEach(func() {
		if input.Teardown != nil {
			Ω.Expect(input.Teardown(knownTopic)).To(Ω.Succeed())
		}
	})

	ginkgo.It("should subscribe", func() {
		Ω.Expect(subject.Subscribe(ctx, knownTopic, handler)).To(Ω.Succeed())
		Ω.Expect(extractDatas(handler.Received)).To(Ω.Equal(extractDatas(input.SeededMessages)))
	})

	ginkgo.It("should stop on bps.Done", func() {
		// allow error wrapping:
		handler.Err = fmt.Errorf("wrapped %w", bps.Done)

		Ω.Expect(subject.Subscribe(ctx, knownTopic, handler)).To(Ω.Succeed())

		// only first message handled - before error is returned:
		Ω.Expect(extractDatas(handler.Received)).To(Ω.Equal(extractDatas(input.SeededMessages[0:1])))
	})

	ginkgo.It("should return handler error", func() {
		expectedErr := errors.New("foo")
		handler.Err = expectedErr

		actualErr := subject.Subscribe(ctx, knownTopic, handler)

		// allow error wrapping:
		Ω.Expect(errors.Is(actualErr, expectedErr)).To(Ω.BeTrue())

		// only first message handled - before error is returned:
		Ω.Expect(extractDatas(handler.Received)).To(Ω.Equal(extractDatas(input.SeededMessages[0:1])))
	})

	ginkgo.It("should stop when context is cancelled", func() {
		cancellableCtx, cancel := context.WithCancel(ctx)
		handler.AfterHandle = cancel

		err := subject.Subscribe(cancellableCtx, knownTopic, handler)

		// allow error wrapping:
		Ω.Expect(errors.Is(err, context.Canceled)).To(Ω.BeTrue())

		// only first message handled - before context is cancelled:
		Ω.Expect(extractDatas(handler.Received)).To(Ω.Equal(extractDatas(input.SeededMessages[0:1])))
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

func extractDatas(msgs []bps.SubMessage) [][]byte {
	datas := make([][]byte, 0, len(msgs))
	for _, msg := range msgs {
		datas = append(datas, msg.Data())
	}
	return datas
}
