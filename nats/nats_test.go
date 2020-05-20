package nats_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bsm/bps"
	"github.com/bsm/bps/internal/lint"
	"github.com/bsm/bps/nats"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	// clusterID holds (default) nats-streaming cluster ID: https://hub.docker.com/_/nats-streaming
	clusterID = "test-cluster"
	stanURL   = "nats://localhost:4222/" + clusterID + "?start_at=first"
)

var _ = Describe("Publisher", func() {
	var subject bps.Publisher
	var ctx = context.Background()

	BeforeEach(func() {
		var err error
		subject, err = bps.NewPublisher(ctx, fmt.Sprintf("%s&%s", stanURL, genClientID()))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
		Expect(subject).To(BeAssignableToTypeOf(&nats.Conn{}))
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
		subject, err = bps.NewSubscriber(ctx, fmt.Sprintf("%s&%s", stanURL, genClientID()))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
		Expect(subject).To(BeAssignableToTypeOf(&nats.Conn{}))
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
	if err := sandboxCheck(); err != nil {
		t.Skipf("skipping test, no sandbox access: %v", err)
		return
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/nats")
}

func sandboxCheck() error {
	conn, err := stan.Connect(clusterID, genClientID())
	if err != nil {
		return err
	}
	return conn.Close()
}

func genClientID() string {
	return fmt.Sprintf("bps-unittest-%d", time.Now().UnixNano())
}

func readMessages(topic string, count int) ([]*bps.PubMessage, error) {
	conn, err := stan.Connect(clusterID, genClientID())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// closed when `count` messages are consumed:
	var done = make(chan struct{})
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
	}, stan.StartAt(pb.StartPosition_First))
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	// wait till messages consumed:
	<-done

	return msgs, nil
}

func seedMessages(topic string, messages []bps.SubMessage) error {
	conn, err := stan.Connect(clusterID, genClientID())
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
