package lint

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/bps"
	"github.com/onsi/ginkgo"
	Ω "github.com/onsi/gomega"
)

// SubscriberInput for the shared test.
type SubscriberInput struct {
	// Subject should return a subject to be tested, with given topic/messages seeded.
	// Caller should handle all the subject teardown/cleanup.
	// Topics are guaranteed to be prefixed with "bps-unittest-".
	Subject func(topic string, messages []bps.SubMessage) bps.Subscriber
}

// Subscriber lints subscribers.
func Subscriber(input *SubscriberInput) {
	var subject bps.Subscriber
	var handler *mockHandler
	var topic string

	var ctx context.Context
	var cancel context.CancelFunc

	ginkgo.BeforeEach(func() {
		cycle := time.Now().UnixNano()
		topic = fmt.Sprintf("bps-unittest-topic-%d-a", cycle)

		subject = input.Subject(topic, []bps.SubMessage{
			bps.RawSubMessage("message-1"),
			bps.RawSubMessage("message-2"),
		})
		handler = &mockHandler{}

		// give consumer max 5s per test:
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	})

	ginkgo.AfterEach(func() {
		cancel()
	})

	ginkgo.It("should subscribe", func() {
		go func() {
			defer ginkgo.GinkgoRecover()

			Ω.Expect(subject.Subscribe(ctx, topic, handler, bps.StartAt(bps.PositionOldest))).To(Ω.Or(
				Ω.Succeed(),
				Ω.MatchError(context.Canceled),
			))
		}()

		Ω.Eventually(handler.Len).Should(Ω.Equal(2))
		Ω.Expect(handler.Data()).To(Ω.ConsistOf("message-1", "message-2"))
	})

	ginkgo.It("should stop on bps.Done", func() {
		// allow error wrapping:
		handler.Err = fmt.Errorf("wrapped %w", bps.Done)

		Ω.Expect(subject.Subscribe(ctx, topic, handler, bps.StartAt(bps.PositionOldest))).To(Ω.Succeed())

		// only one (first received) message handled before error is returned:
		Ω.Expect(handler.Len()).To(Ω.Equal(1))
	})

	ginkgo.It("should return handler error", func() {
		expectedErr := errors.New("foo")
		handler.Err = expectedErr

		actualErr := subject.Subscribe(ctx, topic, handler, bps.StartAt(bps.PositionOldest))

		// allow error wrapping:
		Ω.Expect(errors.Is(actualErr, expectedErr)).To(Ω.BeTrue())

		// only one (first received) message handled - before error is returned:
		Ω.Expect(handler.Len()).To(Ω.Equal(1))
	})
}

type mockHandler struct {
	mu   sync.RWMutex
	data []string

	Err error // error to return after handling (first) message
}

func (h *mockHandler) Handle(msg bps.SubMessage) error {
	h.mu.Lock()
	h.data = append(h.data, string(msg.Data()))
	h.mu.Unlock()

	return h.Err
}

func (h *mockHandler) Data() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.data
}

func (h *mockHandler) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.data)
}
