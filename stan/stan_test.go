/*
BPS_TEST=stan STAN_ADDRS=127.0.0.1:4223 go test -count=1
*/

package stan_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/bsm/bps"
	"github.com/bsm/bps/internal/lint"
	"github.com/nats-io/stan.go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Publisher", func() {
	var subject bps.Publisher
	var ctx = context.Background()

	BeforeEach(func() {
		var err error
		subject, err = bps.NewPublisher(ctx, fmt.Sprintf("stan://%s/%s?client_id=%s", stanAddrs, clusterID, bps.GenClientID()))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	Context("lint", func() {
		var shared lint.PublisherInput

		BeforeEach(func() {
			shared = lint.PublisherInput{
				Subject:  subject,
				Messages: readMessages,
			}
		})

		lint.PublisherPositionNewest(&shared)
		lint.PublisherPositionOldest(&shared)
	})
})

var _ = Describe("Subscriber", func() {
	var subject bps.Subscriber
	var ctx = context.Background()

	BeforeEach(func() {
		var err error
		subject, err = bps.NewSubscriber(ctx, fmt.Sprintf("stan://%s/%s?client_id=%s", stanAddrs, clusterID, bps.GenClientID()))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	Context("lint", func() {
		var shared lint.SubscriberInput

		BeforeEach(func() {
			shared = lint.SubscriberInput{
				Subject: subject,
				Seed: func(topic string, messages []bps.SubMessage) {
					Expect(seedMessages(topic, messages)).To(Succeed())
				},
			}
		})

		lint.SubscriberPositionNewest(&shared)
		lint.SubscriberPositionOldest(&shared)
	})
})

// Not a proper Queue Subscriptions test, but smth to make sure subscriber is not broken
var _ = Describe("Subscriber (Queue)", func() {
	var subject bps.Subscriber
	var ctx = context.Background()

	BeforeEach(func() {
		clientID := bps.GenClientID()
		queueGroup := bps.GenClientID() // something random-ish for tests as well

		var err error
		subject, err = bps.NewSubscriber(ctx, fmt.Sprintf("stan://%s/%s?client_id=%s&queue_group=%s", stanAddrs, clusterID, clientID, queueGroup))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	Context("lint", func() {
		var shared lint.SubscriberInput

		BeforeEach(func() {
			shared = lint.SubscriberInput{
				Subject: subject,
				Seed: func(topic string, messages []bps.SubMessage) {
					Expect(seedMessages(topic, messages)).To(Succeed())
				},
			}
		})

		lint.SubscriberPositionNewest(&shared)
		lint.SubscriberPositionOldest(&shared)
	})
})

// ----------------------------------------------------------------------------

func TestSuite(t *testing.T) {
	if !strings.Contains(os.Getenv("BPS_TEST"), "stan") {
		t.Skipf("skipping test")
		return
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/stan")
}

func readMessages(topic string, count int) ([]*bps.PubMessage, error) {
	conn, err := stan.Connect(clusterID, bps.GenClientID(), stan.NatsURL("nats://"+stanAddrs))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// closed when `count` messages are consumed:
	var done = make(chan struct{}, 1)
	defer close(done)

	var msgs []*bps.PubMessage
	var sub stan.Subscription

	sub, err = conn.Subscribe(topic, func(msg *stan.Msg) {
		msgs = append(msgs, &bps.PubMessage{
			Data: msg.Data,
		})
		if len(msgs) >= count {
			select {
			case done <- struct{}{}:
			default:
			}
		}
	}, stan.DeliverAllAvailable())
	if err != nil {
		return nil, err
	}
	defer func() { _ = sub.Unsubscribe() }()

	// wait till messages consumed:
	<-done

	return msgs, nil
}

func seedMessages(topic string, messages []bps.SubMessage) error {
	conn, err := stan.Connect(clusterID, bps.GenClientID(), stan.NatsURL("nats://"+stanAddrs))
	if err != nil {
		return err
	}
	defer conn.Close()

	for i, msg := range messages {
		if err := conn.Publish(topic, msg.Data()); err != nil {
			return fmt.Errorf("seed %d of %d: %w", i, len(messages), err)
		}
	}
	return nil
}
