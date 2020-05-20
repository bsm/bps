package kafka_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/bps"
	"github.com/bsm/bps/internal/lint"
	"github.com/bsm/bps/kafka"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var brokerAddrs = []string{"127.0.0.1:9092", "127.0.0.1:9093", "127.0.0.1:9094"}

var _ = Describe("Publisher", func() {
	var subject *kafka.Publisher
	var _ bps.Publisher = subject
	var ctx = context.Background()

	BeforeEach(func() {
		pub, err := bps.NewPublisher(ctx, "kafka://"+strings.Join(brokerAddrs, ",")+"?flush.messages=1&flush.frequency=1ms")
		Expect(err).NotTo(HaveOccurred())
		subject = pub.(*kafka.Publisher)
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

var _ = Describe("SyncPublisher", func() {
	var subject *kafka.SyncPublisher
	var _ bps.Publisher = subject
	var ctx = context.Background()

	BeforeEach(func() {
		pub, err := bps.NewPublisher(ctx, "kafka+sync://"+strings.Join(brokerAddrs, ","))
		Expect(err).NotTo(HaveOccurred())
		subject = pub.(*kafka.SyncPublisher)
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
	var subject *kafka.Subscriber
	var _ bps.Subscriber = subject
	var ctx = context.Background()

	BeforeEach(func() {
		pub, err := bps.NewSubscriber(ctx, "kafka://"+strings.Join(brokerAddrs, ",")+"?offsets.initial=oldest")
		Expect(err).NotTo(HaveOccurred())
		subject = pub.(*kafka.Subscriber)
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

// ------------------------------------------------------------------------

func TestSuite(t *testing.T) {
	if err := sandboxCheck(); err != nil {
		t.Skipf("skipping test, no sandbox access: %v", err)
		return
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/kafka")
}

func sandboxCheck() error {
	pub, err := kafka.NewPublisher(brokerAddrs, nil)
	if err != nil {
		return err
	}
	return pub.Close()
}

func readMessages(topic string, _ int) ([]*bps.PubMessage, error) {
	csmr, err := sarama.NewConsumer(brokerAddrs, nil)
	if err != nil {
		return nil, err
	}
	defer csmr.Close()

	parts, err := csmr.Partitions(topic)
	if err != nil {
		return nil, err
	}

	var msgs []*bps.PubMessage
	for _, part := range parts {
		pc, err := csmr.ConsumePartition(topic, part, sarama.OffsetOldest)
		if err != nil {
			return nil, err
		}
		msgs = readPartition(msgs, pc)
	}
	return msgs, nil
}

func readPartition(msgs []*bps.PubMessage, pc sarama.PartitionConsumer) []*bps.PubMessage {
	defer pc.Close()

	for {
		select {
		case msg := <-pc.Messages():
			msgs = append(msgs, &bps.PubMessage{ID: string(msg.Key), Data: msg.Value})
		case <-time.After(5 * time.Millisecond):
			return msgs
		}
	}
}

func seedMessages(topic string, messages []bps.SubMessage) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	pr, err := sarama.NewSyncProducer(brokerAddrs, config)
	if err != nil {
		return err
	}
	defer pr.Close()

	for _, msg := range messages {
		_, _, err := pr.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(msg.Data()),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
