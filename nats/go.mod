module github.com/bsm/bps/nats

go 1.16

replace github.com/bsm/bps => ../

require (
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/bsm/ginkgo v1.16.5
	github.com/bsm/gomega v1.17.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/nats-io/nats-server/v2 v2.6.6 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20211122170419-d7c1d78a50fc
	google.golang.org/protobuf v1.27.1 // indirect
)
