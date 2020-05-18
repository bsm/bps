package file_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bsm/bps"
	"github.com/bsm/bps/internal/lint"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/bsm/bps/file"
)

var _ = Describe("Publisher", func() {
	var subject bps.Publisher
	var ctx = context.Background()
	var dir string

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "bps-file-test")
		Expect(err).NotTo(HaveOccurred())

		subject, err = bps.NewPublisher(ctx, "file://"+dir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	Context("lint", func() {
		var shared lint.PublisherInput

		readMessages := func(topic string, _ int) ([]*bps.PubMessage, error) {
			f, err := os.Open(filepath.Join(dir, topic))
			if err != nil {
				return nil, err
			}
			defer f.Close()

			var res []*bps.PubMessage
			for dec := json.NewDecoder(f); dec.More(); {
				var msg *bps.PubMessage
				if err := dec.Decode(&msg); err != nil {
					return nil, err
				}
				res = append(res, msg)
			}
			return res, nil
		}

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
	var dir string

	BeforeEach(func() {
		var err error

		dir, err = ioutil.TempDir("", "bps-file-test")
		Expect(err).NotTo(HaveOccurred())

		subject, err = bps.NewSubscriber(ctx, "file://"+dir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	Context("lint", func() {
		var shared lint.SubscriberInput

		BeforeEach(func() {
			shared = lint.SubscriberInput{
				Subject: subject,
				SeededMessages: []bps.SubMessage{
					bps.RawSubMessage(`{"msg":1}`),
					bps.RawSubMessage(`{"msg":2}`),
				},
				Setup: func(topic string) error {
					return ioutil.WriteFile(filepath.Join(dir, topic), []byte(`
						{"msg":1}
						{"msg":2}
					`), 0666)
				},
				// no need for Teardown as entire tmp dir will be removed
			}
		})

		lint.Subscriber(&shared)
	})
})

// ------------------------------------------------------------------------

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/file")
}
