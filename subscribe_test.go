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
	var subject *bps.InMemSubscriber

	ctx := context.Background()

	noop := bps.HandlerFunc(func(bps.SubMessage) error { return nil })

	BeforeEach(func() {
		subject = bps.NewInMemSubscriber(
			map[string][]bps.SubMessage{
				"foo": []bps.SubMessage{
					bps.RawSubMessage("foo1"),
					bps.RawSubMessage("foo2"),
				},
			},
		)
	})

	AfterEach(func() {
		_ = subject.Close()
	})

	Describe("Subscribe", func() {
		It("should handle messages", func() {
			var handled []bps.SubMessage
			handler := bps.HandlerFunc(func(msg bps.SubMessage) error {
				handled = append(handled, msg)
				return nil
			})

			Expect(subject.Subscribe(ctx, "foo", handler)).To(Succeed())

			Expect(handled).To(Equal([]bps.SubMessage{
				bps.RawSubMessage("foo1"),
				bps.RawSubMessage("foo2"),
			}))
		})

		It("should not fail on unknown topic", func() {
			Expect(subject.Subscribe(ctx, "bar", noop)).To(Succeed())
		})

		It("should return handler error", func() {
			err := errors.New("oops")
			handler := bps.HandlerFunc(func(bps.SubMessage) error { return err })

			Expect(subject.Subscribe(ctx, "foo", handler)).To(MatchError(err))
		})

		It("should check context", func() {
			canceledCtx, cancel := context.WithCancel(context.Background())
			cancel()

			Expect(subject.Subscribe(canceledCtx, "foo", noop)).To(MatchError(context.Canceled))
		})
	})
})
