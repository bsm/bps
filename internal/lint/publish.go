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
	Messages func(string, int) ([]*bps.Message, error)
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
		Ω.Expect(subject.Topic(topicA).Publish(ctx, &bps.Message{Data: []byte("v1")})).To(Ω.Succeed())
		Ω.Expect(subject.Topic(topicB).Publish(ctx, &bps.Message{Data: []byte("v2")})).To(Ω.Succeed())
		Ω.Expect(subject.Topic(topicA).Publish(ctx, &bps.Message{Data: []byte("v3")})).To(Ω.Succeed())

		Ω.Eventually(func() ([]*bps.Message, error) {
			return input.Messages(topicA, 2)
		}, "3s").Should(haveData("v1", "v3"))

		Ω.Eventually(func() ([]*bps.Message, error) {
			return input.Messages(topicB, 1)
		}, "3s").Should(haveData("v2"))
	})

	ginkgo.It("should publish batches", func() {
		Ω.Expect(subject.Topic(topicA).PublishBatch(ctx, []*bps.Message{
			{Data: []byte("v1")},
			{Data: []byte("v3")},
			{Data: []byte("v5")},
			{Data: []byte("v7")},
			{Data: []byte("v9")},
		})).To(Ω.Succeed())

		Ω.Expect(subject.Topic(topicB).PublishBatch(ctx, []*bps.Message{
			{Data: []byte("v2")},
			{Data: []byte("v4")},
		})).To(Ω.Succeed())

		Ω.Expect(subject.Topic(topicB).PublishBatch(ctx, []*bps.Message{
			{Data: []byte("v6")},
			{Data: []byte("v8")},
		})).To(Ω.Succeed())

		Ω.Eventually(func() ([]*bps.Message, error) {
			return input.Messages(topicA, 5)
		}, "3s").Should(haveData("v1", "v3", "v5", "v7", "v9"))

		Ω.Eventually(func() ([]*bps.Message, error) {
			return input.Messages(topicB, 4)
		}, "3s").Should(haveData("v2", "v4", "v6", "v8"))
	})
}

func haveData(vals ...string) types.GomegaMatcher {
	return Ω.WithTransform(func(msgs []*bps.Message) []string {
		var strs []string
		for _, m := range msgs {
			strs = append(strs, string(m.Data))
		}
		return strs
	}, Ω.ConsistOf(vals))
}
