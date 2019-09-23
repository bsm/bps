package lint

import (
	"context"

	"github.com/bsm/bps"
	"github.com/onsi/ginkgo"
	Ω "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// PublisherInput for the shared test.
type PublisherInput struct {
	Subject  bps.Publisher
	Messages func(string) ([]*bps.Message, error)
}

// Publisher lints publishers.
func Publisher(input *PublisherInput) {
	var subject bps.Publisher
	var ctx = context.Background()

	ginkgo.BeforeEach(func() {
		subject = input.Subject
	})

	ginkgo.It("should publish", func() {
		Ω.Expect(subject.Topic("topicA").Publish(ctx, &bps.Message{Data: []byte("v1")})).To(Ω.Succeed())
		Ω.Expect(subject.Topic("topicB").Publish(ctx, &bps.Message{Data: []byte("v2")})).To(Ω.Succeed())
		Ω.Expect(subject.Topic("topicA").Publish(ctx, &bps.Message{Data: []byte("v3")})).To(Ω.Succeed())

		Ω.Expect(input.Messages("topicA")).To(haveMessageData("v1", "v3"))
		Ω.Expect(input.Messages("topicB")).To(haveMessageData("v2"))
	})

	ginkgo.It("should publish batches", func() {
		Ω.Expect(subject.Topic("topicA").PublishBatch(ctx, []*bps.Message{
			{Data: []byte("v1")},
			{Data: []byte("v3")},
			{Data: []byte("v5")},
			{Data: []byte("v7")},
		})).To(Ω.Succeed())
		Ω.Expect(subject.Topic("topicB").PublishBatch(ctx, []*bps.Message{
			{Data: []byte("v2")},
			{Data: []byte("v4")},
		})).To(Ω.Succeed())

		Ω.Expect(input.Messages("topicA")).To(haveMessageData("v1", "v3", "v5", "v7"))
		Ω.Expect(input.Messages("topicB")).To(haveMessageData("v2", "v4"))
	})
}

func haveMessageData(vals ...string) types.GomegaMatcher {
	return Ω.WithTransform(func(msgs []*bps.Message) []string {
		var strs []string
		for _, m := range msgs {
			strs = append(strs, string(m.Data))
		}
		return strs
	}, Ω.Equal(vals))
}
