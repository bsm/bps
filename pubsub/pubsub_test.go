package pubsub_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	native "cloud.google.com/go/pubsub"
	"github.com/bsm/bps"
	"github.com/bsm/bps/internal/lint"
	"github.com/bsm/bps/pubsub"
	"google.golang.org/api/iterator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

		lint.Publisher(&shared)
	})
})

// ------------------------------------------------------------------------

const projectID = "bsm-tech"

func TestSuite(t *testing.T) {
	if err := sandboxCheck(); err != nil {
		t.Skipf("skipping test, no sandbox access: %v", err)
		return
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/pubsub")
}

func sandboxCheck() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	psc, err := native.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer psc.Close()

	topic, err := psc.CreateTopic(ctx, fmt.Sprintf("bps-unittest-ping-%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	defer topic.Delete(ctx)

	return nil
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

	return nil
}
