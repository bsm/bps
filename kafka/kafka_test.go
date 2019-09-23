package kafka_test

import (
	"context"
	"testing"

	"github.com/bsm/bps"
	"github.com/bsm/bps/kafka"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = PDescribe("Publisher", func() {
	var subject *kafka.Publisher
	var _ bps.Publisher = subject
	var ctx = context.Background()

	BeforeEach(func() {
		pub, err := bps.NewPublisher(ctx, "kafka://127.0.0.1:9092,127.0.0.1:9093,127.0.0.1:9094/")
		Expect(err).NotTo(HaveOccurred())
		subject = pub.(*kafka.Publisher)
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	PIt("should lint", func() {
	})
})

var _ = PDescribe("SyncPublisher", func() {
	var subject *kafka.SyncPublisher
	var _ bps.Publisher = subject
	var ctx = context.Background()

	BeforeEach(func() {
		pub, err := bps.NewPublisher(ctx, "kafka+sync://127.0.0.1:9092,127.0.0.1:9093,127.0.0.1:9094/")
		Expect(err).NotTo(HaveOccurred())
		subject = pub.(*kafka.SyncPublisher)
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	PIt("should lint", func() {
	})
})

// ------------------------------------------------------------------------

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/kafka")
}
