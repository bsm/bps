package bps_test

import (
	"context"
	"net/url"

	"github.com/bsm/bps"
	"github.com/bsm/bps/internal/lint"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

func init() {
	bps.RegisterPublisher("mem", func(_ context.Context, u *url.URL) (bps.Publisher, error) {
		return bps.NewInMemPublisher(), nil
	})
}

var _ = Describe("RegisterPublisher", func() {
	var ctx = context.Background()

	It("should panic when trying to register the same scheme 2x", func() {
		Expect(func() {
			bps.RegisterPublisher("mem", nil)
		}).To(Panic())
	})

	It("should allow to init publishers by URL", func() {
		pub, err := bps.NewPublisher(ctx, "mem://test.host/path")
		Expect(err).NotTo(HaveOccurred())
		Expect(pub).To(BeAssignableToTypeOf(&bps.InMemPublisher{}))
		Expect(pub.Close()).To(Succeed())
	})

	It("should fail on unknown schemes", func() {
		_, err := bps.NewPublisher(ctx, "unknown://test.host/path")
		Expect(err).To(MatchError(`unknown URL scheme "unknown"`))
	})
})

var _ = Describe("InMemPublisher", func() {
	var subject *bps.InMemPublisher

	BeforeEach(func() {
		subject = bps.NewInMemPublisher()
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	Context("lint", func() {
		var shared lint.PublisherInput

		readMessages := func(topic string, _ int) ([]*bps.PubMessage, error) {
			return subject.Topic(topic).(*bps.InMemPubTopic).Messages(), nil
		}

		BeforeEach(func() {
			subject = bps.NewInMemPublisher()
			shared = lint.PublisherInput{
				Subject:  subject,
				Messages: readMessages,
			}
		})

		lint.PublisherPositionOldest(&shared)
	})
})
