module github.com/bsm/bps/kafka

go 1.13

replace github.com/bsm/bps => ../

require (
	github.com/Shopify/sarama v1.26.4
	github.com/bsm/bps v0.0.0-20200605133303-c9abd8b1e2f6
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
)
