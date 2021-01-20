package lint

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/bps"
	"github.com/onsi/ginkgo"
	Ω "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// PublisherInput for the shared test.
type PublisherInput struct {
	Subject  bps.Publisher
	Messages func(string, int) ([]*bps.PubMessage, error)
	Setup    func(topics ...string) error
	Teardown func(topics ...string) error
}

// PublisherPositionNewest lints publishers for StartAt=PositionNewest.
// It does subscribe and then publish.
func PublisherPositionNewest(input *PublisherInput) {
	linter := &publisherPositionNewestLinter{
		publisherLinter: &publisherLinter{
			ctx:   context.Background(),
			input: input,
		},
	}

	ginkgo.BeforeEach(linter.Prepare)
	ginkgo.AfterEach(linter.Cleanup)

	ginkgo.Context("StartAt=PositionNewest", func() {
		ginkgo.It("should publish", linter.Lint)
	})
}

// PublisherPositionOldest lints publishers for StartAt=PositionOldest.
// It does publish and then subscribe.
func PublisherPositionOldest(input *PublisherInput) {
	linter := publisherPositionOldestLinter{
		publisherLinter: &publisherLinter{
			ctx:   context.Background(),
			input: input,
		},
	}

	ginkgo.BeforeEach(linter.Prepare)
	ginkgo.AfterEach(linter.Cleanup)

	ginkgo.Context("StartAt=PositionOldest", func() {
		ginkgo.It("should publish", linter.Lint)
	})
}

// ----------------------------------------------------------------------------

type publisherLinter struct {
	ctx            context.Context
	input          *PublisherInput
	topicA, topicB string
}

func (l *publisherLinter) Prepare() {
	cycle := time.Now().UnixNano()
	l.topicA = fmt.Sprintf("bps-unittest-topic-%d-a", cycle)
	l.topicB = fmt.Sprintf("bps-unittest-topic-%d-b", cycle)

	if l.input.Setup != nil {
		Ω.Expect(l.input.Setup(l.topicA, l.topicB)).To(Ω.Succeed())
	}
}

func (l *publisherLinter) Cleanup() {
	if l.input.Teardown != nil {
		Ω.Expect(l.input.Teardown(l.topicA, l.topicB)).To(Ω.Succeed())
	}
}

// ----------------------------------------------------------------------------

type publisherPositionNewestLinter struct {
	*publisherLinter
}

func (l *publisherPositionNewestLinter) Lint() {
	var subscribers sync.WaitGroup

	subscribers.Add(1)
	go func() {
		defer ginkgo.GinkgoRecover()
		defer subscribers.Done()

		Ω.Eventually(func() ([]*bps.PubMessage, error) {
			return l.input.Messages(l.topicA, 2)
		}, 3*subscriptionWaitDelay).Should(haveData("v1", "v3"))
	}()

	subscribers.Add(1)
	go func() {
		defer ginkgo.GinkgoRecover()
		defer subscribers.Done()

		Ω.Eventually(func() ([]*bps.PubMessage, error) {
			return l.input.Messages(l.topicB, 1)
		}, 3*subscriptionWaitDelay).Should(haveData("v2"))
	}()

	// give subscribers some time to actually start consuming:
	time.Sleep(subscriptionWaitDelay)

	Ω.Expect(l.input.Subject.Topic(l.topicA).Publish(l.ctx, &bps.PubMessage{Data: []byte("v1")})).To(Ω.Succeed())
	Ω.Expect(l.input.Subject.Topic(l.topicB).Publish(l.ctx, &bps.PubMessage{Data: []byte("v2")})).To(Ω.Succeed())
	Ω.Expect(l.input.Subject.Topic(l.topicA).Publish(l.ctx, &bps.PubMessage{Data: []byte("v3")})).To(Ω.Succeed())

	subscribers.Wait()
}

// ----------------------------------------------------------------------------

type publisherPositionOldestLinter struct {
	*publisherLinter
}

func (l *publisherPositionOldestLinter) Lint() {
	Ω.Expect(l.input.Subject.Topic(l.topicA).Publish(l.ctx, &bps.PubMessage{Data: []byte("v1")})).To(Ω.Succeed())
	Ω.Expect(l.input.Subject.Topic(l.topicB).Publish(l.ctx, &bps.PubMessage{Data: []byte("v2")})).To(Ω.Succeed())
	Ω.Expect(l.input.Subject.Topic(l.topicA).Publish(l.ctx, &bps.PubMessage{Data: []byte("v3")})).To(Ω.Succeed())

	Ω.Eventually(func() ([]*bps.PubMessage, error) {
		return l.input.Messages(l.topicA, 2)
	}, 3*subscriptionWaitDelay).Should(haveData("v1", "v3"))

	Ω.Eventually(func() ([]*bps.PubMessage, error) {
		return l.input.Messages(l.topicB, 1)
	}, 3*subscriptionWaitDelay).Should(haveData("v2"))
}

func haveData(vals ...string) types.GomegaMatcher {
	return Ω.WithTransform(func(msgs []*bps.PubMessage) []string {
		var strs []string
		for _, m := range msgs {
			strs = append(strs, string(m.Data))
		}
		return strs
	}, Ω.ConsistOf(vals))
}
