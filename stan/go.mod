module github.com/bsm/bps/stan

go 1.15

replace github.com/bsm/bps => ../

require (
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/nats-io/nats-streaming-server v0.20.0 // indirect
	github.com/nats-io/nats.go v1.10.0
	github.com/nats-io/stan.go v0.8.2
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
)
