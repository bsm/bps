module github.com/bsm/bps/stan

go 1.17

replace github.com/bsm/bps => ../

require (
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/bsm/ginkgo v1.16.5
	github.com/bsm/gomega v1.17.0
	github.com/nats-io/nats.go v1.13.1-0.20220121202836-972a071d373d
	github.com/nats-io/stan.go v0.10.2
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/nats-io/nats-server/v2 v2.7.2 // indirect
	github.com/nats-io/nats-streaming-server v0.24.1 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
