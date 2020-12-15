package nats_test

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

const (
	// clusterID holds (default) nats-streaming cluster ID: https://hub.docker.com/_/nats-streaming
	clusterID = "test-cluster"
)

var _ = Describe("Publisher", func() {
	var subject bps.Publisher
	var ctx = context.Background()

	BeforeEach(func() {
		var err error
		subject, err = bps.NewPublisher(ctx, fmt.Sprintf("%s?client_id=%s", "nats://"+natsAddr+"/"+clusterID, bps.GenClientID()))
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

		lint.Publisher(&shared)
	})
})

var _ = Describe("Subscriber", func() {
	var subject bps.Subscriber
	var ctx = context.Background()

	BeforeEach(func() {
		var err error
		subject, err = bps.NewSubscriber(ctx, fmt.Sprintf("%s?client_id=%s", "nats://"+natsAddr+"/"+clusterID, bps.GenClientID()))
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
				Subject: func(topic string, messages []bps.SubMessage) bps.Subscriber {
					Expect(seedMessages(topic, messages)).To(Succeed())
					return subject
				},
			}
		})

		lint.Subscriber(&shared)
	})
})

// ----------------------------------------------------------------------------

func TestSuite(t *testing.T) {
	if !strings.Contains(os.Getenv("BPS_TEST"), "nats") {
		t.Skipf("skipping test")
		return
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/nats")
}

func readMessages(topic string, count int) ([]*bps.PubMessage, error) {
	conn, err := stan.Connect(clusterID, bps.GenClientID())
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
	defer sub.Unsubscribe()

	// wait till messages consumed:
	<-done

	return msgs, nil
}

func seedMessages(topic string, messages []bps.SubMessage) error {
	conn, err := stan.Connect(clusterID, bps.GenClientID())
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
