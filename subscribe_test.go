package bps_test

import (
	"context"
	"errors"
	"net/url"

	"github.com/bsm/bps"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	bps.RegisterSubscriber("mem", func(_ context.Context, u *url.URL) (bps.Subscriber, error) {
		return dummySubscriber{}, nil
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
		Expect(sub).To(BeAssignableToTypeOf(dummySubscriber{}))
	})

	It("should fail on unknown schemes", func() {
		_, err := bps.NewSubscriber(ctx, "unknown://test.host/path")
		Expect(err).To(MatchError(`unknown URL scheme "unknown"`))
	})
})

var _ = Describe("SubscribeOptions", func() {
	It("should normalize / apply defaults", func() {
		Expect((*bps.SubscribeOptions)(nil).Normalize()).To(Equal(&bps.SubscribeOptions{
			BatchSize: 1,
		}))
		Expect(new(bps.SubscribeOptions).Normalize()).To(Equal(&bps.SubscribeOptions{
			BatchSize: 1,
		}))
	})
})

// ----------------------------------------------------------------------------

type dummySubscriber struct{}

func (dummySubscriber) Subscribe(context.Context, string, func([]bps.Message) error, *bps.SubscribeOptions) error {
	return errors.New("not expected to be called")
}

func (dummySubscriber) Close() error {
	return errors.New("not expected to be called")
}
