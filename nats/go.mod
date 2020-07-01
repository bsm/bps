module github.com/bsm/bps/nats

go 1.14

replace github.com/bsm/bps => ../

require (
	github.com/bsm/bps v0.0.0-20200605133303-c9abd8b1e2f6
	github.com/nats-io/nats-streaming-server v0.18.0 // indirect
	github.com/nats-io/stan.go v0.7.0
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
)
