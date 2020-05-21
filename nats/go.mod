module github.com/bsm/bps/nats

go 1.14

replace github.com/bsm/bps => ../

require (
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/nats-io/nats-server/v2 v2.1.7 // indirect
	github.com/nats-io/nats-streaming-server v0.17.0 // indirect
	github.com/nats-io/stan.go v0.6.0
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.0
)
