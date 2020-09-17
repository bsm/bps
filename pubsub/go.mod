module github.com/bsm/bps/pubsub

go 1.14

replace github.com/bsm/bps => ../

require (
	cloud.google.com/go/pubsub v1.6.1
	github.com/bsm/bps v0.1.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	google.golang.org/api v0.32.0
)
