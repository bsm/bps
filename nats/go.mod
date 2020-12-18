module github.com/bsm/bps/nats

go 1.14

replace github.com/bsm/bps => ../

require (
	github.com/bsm/bps v0.1.0
	github.com/nats-io/nats.go v1.10.0
	github.com/nats-io/stan.go v0.7.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	go.uber.org/multierr v1.6.0
)
