module github.com/bsm/bps/pubsub

go 1.14

replace github.com/bsm/bps => ../

require (
	cloud.google.com/go/pubsub v1.6.1
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.1.2 // indirect
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	google.golang.org/api v0.31.0
)
