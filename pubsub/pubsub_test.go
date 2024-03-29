package pubsub_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	native "cloud.google.com/go/pubsub"
	"github.com/bsm/bps"
	"github.com/bsm/bps/internal/lint"
	"github.com/bsm/bps/pubsub"
	"google.golang.org/api/iterator"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

var _ = Describe("Publisher", func() {
	var subject *pubsub.Publisher
	var _ bps.Publisher = subject
	var ctx = context.Background()

	BeforeEach(func() {
		pub, err := bps.NewPublisher(ctx, "pubsub://"+projectID)
		Expect(err).NotTo(HaveOccurred())

		subject = pub.(*pubsub.Publisher)
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
				Setup:    setupCB,
				Teardown: teardownCB,
			}
		})

		// TODO: check how it actually works, there's no support for StartAt variants:
		lint.PublisherPositionNewest(&shared)
		lint.PublisherPositionOldest(&shared)
	})
})

var _ = Describe("Subscriber", func() {
	var subject *pubsub.Subscriber
	var _ bps.Subscriber = subject
	var ctx = context.Background()

	BeforeEach(func() {
		pub, err := bps.NewSubscriber(ctx, "pubsub://"+projectID)
		Expect(err).NotTo(HaveOccurred())

		subject = pub.(*pubsub.Subscriber)
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())

		// wipes all the "bps-unittest-*" topics;
		// subscriptions are expected to be deleted by Subscriber implementation itself:
		Expect(teardownCB()).To(Succeed())
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

		// TODO: check how it actually works, there's no support for StartAt variants:
		lint.SubscriberPositionOldest(&shared)
		lint.SubscriberPositionNewest(&shared)
	})
})

// ------------------------------------------------------------------------

const projectID = "bsm-tech"

func TestSuite(t *testing.T) {
	if !strings.Contains(os.Getenv("BPS_TEST"), "pubsub") {
		t.Skipf("skipping test")
		return
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/pubsub")
}

func readMessages(topic string, max int) ([]*bps.PubMessage, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancel()

	psc, err := native.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer psc.Close()

	var msgs []*bps.PubMessage
	var mu sync.Mutex

	if err := psc.Subscription(topic).Receive(ctx, func(ctx context.Context, msg *native.Message) {
		msg.Ack()

		mu.Lock()
		defer mu.Unlock()

		msgs = append(msgs, &bps.PubMessage{ID: msg.ID, Data: msg.Data})
		if len(msgs) == max {
			cancel()
		}
	}); err != nil {
		return nil, err
	}
	return msgs, nil
}

func setupCB(topics ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	psc, err := native.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer psc.Close()

	for _, topic := range topics {
		t, err := psc.CreateTopic(ctx, topic)
		if err != nil {
			return err
		}
		if _, err := psc.CreateSubscription(ctx, topic, native.SubscriptionConfig{Topic: t}); err != nil {
			return err
		}
	}
	return nil
}

func teardownCB(...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	psc, err := native.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer psc.Close()

	for it := psc.Topics(ctx); true; {
		t, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}

		if strings.HasPrefix(t.ID(), "bps-unittest-") {
			if err := t.Delete(ctx); err != nil {
				return err
			}
		}
	}

	for it := psc.Subscriptions(ctx); true; {
		s, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}

		if strings.HasPrefix(s.ID(), "bps-unittest-") {
			if err := s.Delete(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func seedMessages(topicName string, messages []bps.SubMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := native.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	topic, err := client.CreateTopic(ctx, topicName)
	if err != nil {
		return err
	}
	defer topic.Stop()

	for i, msg := range messages {
		nativeMsg := &native.Message{
			Data: msg.Data(),
		}
		if _, err := topic.Publish(ctx, nativeMsg).Get(ctx); err != nil {
			return fmt.Errorf("failed to seed %d of %d", i, len(messages))
		}
	}
	return nil
}
