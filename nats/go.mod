module github.com/bsm/bps/nats

go 1.17

replace github.com/bsm/bps => ../

require (
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/bsm/ginkgo v1.16.5
	github.com/bsm/gomega v1.17.0
	github.com/nats-io/nats.go v1.13.1-0.20220308171302-2f2f6968e98d
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/nats-io/nats-server/v2 v2.7.4 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
