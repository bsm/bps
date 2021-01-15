package bps_test

import (
	"context"
	"net/url"

	"github.com/bsm/bps"
	"github.com/bsm/bps/internal/lint"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	bps.RegisterSubscriber("mem", func(_ context.Context, u *url.URL) (bps.Subscriber, error) {
		return bps.NewInMemSubscriber(nil), nil
	})
}

var _ = Describe("RegisterSubscriber", func() {
	var ctx = context.Background()

	It("should panic when trying to register the same scheme 2x", func() {
		Expect(func() {
			bps.RegisterSubscriber("mem", nil)
		}).To(Panic())
	})

	It("should allow to init subscribers by URL", func() {
		sub, err := bps.NewSubscriber(ctx, "mem://test.host/path")
		Expect(err).NotTo(HaveOccurred())
		defer sub.Close()

		Expect(sub).To(BeAssignableToTypeOf(&bps.InMemSubscriber{}))
	})

	It("should fail on unknown schemes", func() {
		_, err := bps.NewSubscriber(ctx, "unknown://test.host/path")
		Expect(err).To(MatchError(`unknown URL scheme "unknown"`))
	})
})

var _ = Describe("InMemSubscriber", func() {
	Context("lint", func() {
		var subject *bps.InMemSubscriber
		var shared lint.SubscriberInput

		BeforeEach(func() {
			subject = bps.NewInMemSubscriber(nil)

			shared = lint.SubscriberInput{
				Subject: subject,
				Seed: func(topic string, messages []bps.SubMessage) {
					subject.Replace(
						map[string][]bps.SubMessage{
							topic: messages,
						},
					)
				},
			}
		})

		lint.SubscriberPositionOldest(&shared)
	})
})
