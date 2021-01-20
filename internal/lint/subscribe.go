package lint

import (
	"fmt"
	"sync"
	"time"

	"github.com/bsm/bps"
	"github.com/onsi/ginkgo"
	Ω "github.com/onsi/gomega"
)

// SubscriberInput for the shared test.
type SubscriberInput struct {
	// Subject should be set to a subject to be tested.
	// Caller should handle all the subject teardown/cleanup.
	Subject bps.Subscriber

	// Seed can seed given topic/messages for Subject unless already seeded.
	// Topics are guaranteed to be prefixed with "bps-unittest-".
	Seed func(topic string, messages []bps.SubMessage)
}

// SubscriberPositionNewest lints subscribers with StartAt=PositionNewest.
// it does subscribe and then publish.
func SubscriberPositionNewest(input *SubscriberInput) {
	linter := &subscriberPositionNewestLinter{
		subscriberLinter: &subscriberLinter{
			input: input,
		},
	}

	ginkgo.BeforeEach(linter.Prepare)

	ginkgo.Context("StartAt=PositionNewest", func() {
		ginkgo.It("should subscribe", linter.Lint)
	})
}

// SubscriberPositionOldest lints subscribers with StartAt=PositionOldest.
// it does subscribe and then publish.
func SubscriberPositionOldest(input *SubscriberInput) {
	linter := &subscriberPositionOldestLinter{
		subscriberLinter: &subscriberLinter{
			input: input,
		},
	}

	ginkgo.BeforeEach(linter.Prepare)

	ginkgo.Context("StartAt=PositionOldest", func() {
		ginkgo.It("should subscribe", linter.Lint)
	})
}

// ----------------------------------------------------------------------------

type subscriberLinter struct {
	input *SubscriberInput

	handler  *mockHandler
	topic    string
	messages []bps.SubMessage
}

func (l *subscriberLinter) Prepare() {
	cycle := time.Now().UnixNano()
	l.topic = fmt.Sprintf("bps-unittest-topic-%d-a", cycle)
	l.messages = []bps.SubMessage{
		bps.RawSubMessage("message-1"),
		bps.RawSubMessage("message-2"),
	}
	l.handler = &mockHandler{}
}

// ----------------------------------------------------------------------------

type subscriberPositionNewestLinter struct {
	*subscriberLinter
}

func (l *subscriberPositionNewestLinter) Lint() {
	// subscribe first:
	sub, err := l.input.Subject.Topic(l.topic).Subscribe(l.handler, bps.StartAt(bps.PositionNewest))
	Ω.Expect(err).NotTo(Ω.HaveOccurred())
	defer sub.Close() // multiple calls must be safe

	// and then publish:
	l.input.Seed(l.topic, l.messages)

	Ω.Eventually(l.handler.Len, 3*subscriptionWaitDelay).Should(Ω.Equal(2))
	Ω.Expect(l.handler.Data()).To(Ω.ConsistOf("message-1", "message-2"))

	Ω.Expect(sub.Close()).To(Ω.Succeed())
}

// ----------------------------------------------------------------------------

type subscriberPositionOldestLinter struct {
	*subscriberLinter
}

func (l *subscriberPositionOldestLinter) Lint() {
	// seed first:
	l.input.Seed(l.topic, l.messages)

	// and then subscribe:
	sub, err := l.input.Subject.Topic(l.topic).Subscribe(l.handler, bps.StartAt(bps.PositionOldest))
	Ω.Expect(err).NotTo(Ω.HaveOccurred())
	defer sub.Close() // multiple calls must be safe

	Ω.Eventually(l.handler.Len, 3*subscriptionWaitDelay).Should(Ω.Equal(2))
	Ω.Expect(l.handler.Data()).To(Ω.ConsistOf("message-1", "message-2"))

	Ω.Expect(sub.Close()).To(Ω.Succeed())
}

// ----------------------------------------------------------------------------

type mockHandler struct {
	mu   sync.RWMutex
	data []string
}

func (h *mockHandler) Handle(msg bps.SubMessage) {
	h.mu.Lock()
	h.data = append(h.data, string(msg.Data()))
	h.mu.Unlock()
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
