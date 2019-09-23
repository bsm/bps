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
	var shared = lint.PublisherInput{Messages: readMessages}

	BeforeEach(func() {
		pub, err := bps.NewPublisher(ctx, "kafka://"+strings.Join(brokerAddrs, ",")+"?flush.messages=1&flush.frequency=1ms")
		Expect(err).NotTo(HaveOccurred())
		subject = pub.(*kafka.Publisher)
		shared.Subject = subject
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	lint.Publisher(&shared)
})

var _ = Describe("SyncPublisher", func() {
	var subject *kafka.SyncPublisher
	var _ bps.Publisher = subject
	var ctx = context.Background()
	var shared = lint.PublisherInput{Messages: readMessages}

	BeforeEach(func() {
		pub, err := bps.NewPublisher(ctx, "kafka+sync://"+strings.Join(brokerAddrs, ","))
		Expect(err).NotTo(HaveOccurred())
		subject = pub.(*kafka.SyncPublisher)
		shared.Subject = subject
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	lint.Publisher(&shared)
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

func readMessages(topic string) ([]*bps.Message, error) {
	csmr, err := sarama.NewConsumer(brokerAddrs, nil)
	if err != nil {
		return nil, err
	}
	defer csmr.Close()

	parts, err := csmr.Partitions(topic)
	if err != nil {
		return nil, err
	}

	var msgs []*bps.Message
	for _, part := range parts {
		pc, err := csmr.ConsumePartition(topic, part, sarama.OffsetOldest)
		if err != nil {
			return nil, err
		}
		msgs = readPartition(msgs, pc)
	}
	return msgs, nil
}

func readPartition(msgs []*bps.Message, pc sarama.PartitionConsumer) []*bps.Message {
	defer pc.Close()

	for {
		select {
		case msg := <-pc.Messages():
			msgs = append(msgs, &bps.Message{ID: string(msg.Key), Data: msg.Value})
		case <-time.After(5 * time.Millisecond):
			return msgs
		}
	}
}
