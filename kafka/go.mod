module github.com/bsm/bps/kafka

go 1.16

replace github.com/bsm/bps => ../

require (
	github.com/Shopify/sarama v1.30.1
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/bsm/ginkgo v1.16.5
	github.com/bsm/gomega v1.17.0
)
