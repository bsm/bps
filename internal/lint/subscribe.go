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

	ginkgo.BeforeEach(func() {
		cycle := time.Now().UnixNano()
		topic = fmt.Sprintf("bps-unittest-topic-%d-a", cycle)

		subject = input.Subject(topic, []bps.SubMessage{
			bps.RawSubMessage("message-1"),
			bps.RawSubMessage("message-2"),
		})
		handler = &mockHandler{}
	})

	ginkgo.It("should subscribe", func() {
		sub, err := subject.Topic(topic).Subscribe(handler, bps.StartAt(bps.PositionOldest))
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		defer sub.Close() // multiple calls are safe

		Ω.Eventually(handler.Len).Should(Ω.Equal(2))
		Ω.Expect(handler.Data()).To(Ω.ConsistOf("message-1", "message-2"))

		Ω.Expect(sub.Close()).To(Ω.Succeed())
	})
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
