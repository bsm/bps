package google_test

import (
	"context"
	"testing"

	"github.com/bsm/bps"
	"github.com/bsm/bps/google"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Publisher", func() {
	var subject *google.Publisher
	var _ bps.Publisher = subject
	var ctx = context.Background()

	BeforeEach(func() {
		pub, err := bps.NewPublisher(ctx, "google://test-project")
		Expect(err).NotTo(HaveOccurred())
		subject = pub.(*google.Publisher)
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
	RunSpecs(t, "bps/google")
}
