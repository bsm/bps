package lint

import (
	"context"
	"fmt"
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

// Publisher lints publishers.
func Publisher(input *PublisherInput) {
	var subject bps.Publisher
	var ctx = context.Background()
	var topicA, topicB string

	ginkgo.BeforeEach(func() {
		subject = input.Subject

		cycle := time.Now().UnixNano()
		topicA = fmt.Sprintf("bps-unittest-topic-%d-a", cycle)
		topicB = fmt.Sprintf("bps-unittest-topic-%d-b", cycle)

		if input.Setup != nil {
			Ω.Expect(input.Setup(topicA, topicB)).To(Ω.Succeed())
		}
	})

	ginkgo.AfterEach(func() {
		if input.Teardown != nil {
			Ω.Expect(input.Teardown(topicA, topicB)).To(Ω.Succeed())
		}
	})

	ginkgo.It("should publish", func() {
		Ω.Expect(subject.Topic(topicA).Publish(ctx, &bps.PubMessage{Data: []byte("v1")})).To(Ω.Succeed())
		Ω.Expect(subject.Topic(topicB).Publish(ctx, &bps.PubMessage{Data: []byte("v2")})).To(Ω.Succeed())
		Ω.Expect(subject.Topic(topicA).Publish(ctx, &bps.PubMessage{Data: []byte("v3")})).To(Ω.Succeed())

		Ω.Eventually(func() ([]*bps.PubMessage, error) {
			return input.Messages(topicA, 2)
		}, "3s").Should(haveData("v1", "v3"))

		Ω.Eventually(func() ([]*bps.PubMessage, error) {
			return input.Messages(topicB, 1)
		}, "3s").Should(haveData("v2"))
	})
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
