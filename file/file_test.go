package file_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/bsm/bps"
	_ "github.com/bsm/bps/file"
	"github.com/bsm/bps/internal/lint"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Publisher", func() {
	var subject bps.Publisher
	var ctx = context.Background()
	var dir string
	var shared = lint.PublisherInput{
		Messages: func(topic string) ([]*bps.Message, error) {
			f, err := os.Open(filepath.Join(dir, topic))
			if err != nil {
				return nil, err
			}
			defer f.Close()

			var res []*bps.Message
			for dec := json.NewDecoder(f); dec.More(); {
				var msg *bps.Message
				if err := dec.Decode(&msg); err != nil {
					return nil, err
				}
				res = append(res, msg)
			}
			return res, nil
		},
	}

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "bps-file-test")

		subject, err = bps.NewPublisher(ctx, "file://"+dir)
		Expect(err).NotTo(HaveOccurred())

		shared.Subject = subject
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	It("should init from URL", func() {
		Expect(subject).NotTo(BeNil())
	})

	lint.Publisher(&shared)
})

// ------------------------------------------------------------------------

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bps/file")
}
