module github.com/bsm/bps/kafka

go 1.14

replace github.com/bsm/bps => ../

require (
	github.com/Shopify/sarama v1.27.0
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
)
