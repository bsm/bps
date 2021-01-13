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
	// It can optionally prepare (pre-create) topic (the same topic will be used for Seed-ing).
	// Caller should handle all the subject teardown/cleanup.
	Subject func(topic string) bps.Subscriber

	// Seed should seed given topic/messages for Subject.
	// Topics are guaranteed to be prefixed with "bps-unittest-".
	Seed func(topic string, messages []bps.SubMessage)
}

// Subscriber lints subscribers.
func Subscriber(input *SubscriberInput) {
	var subject bps.Subscriber
	var handler *mockHandler
	var topic string

	ginkgo.BeforeEach(func() {
		cycle := time.Now().UnixNano()
		topic = fmt.Sprintf("bps-unittest-topic-%d-a", cycle)

		subject = input.Subject(topic)
		handler = &mockHandler{}
	})

	ginkgo.It("should subscribe", func() {
		// produce concurrently:
		go func() {
			defer ginkgo.GinkgoRecover()

			// give subscriber some time to actually start consuming:
			time.Sleep(subscriptionWaitDelay)

			input.Seed(topic, []bps.SubMessage{
				bps.RawSubMessage("message-1"),
				bps.RawSubMessage("message-2"),
			})
		}()

		sub, err := subject.Topic(topic).Subscribe(handler)
		Ω.Expect(err).NotTo(Ω.HaveOccurred())
		defer sub.Close() // multiple calls must be safe

		Ω.Eventually(handler.Len, 3*subscriptionWaitDelay).Should(Ω.Equal(2))
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
