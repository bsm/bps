module github.com/bsm/bps/pubsub

go 1.13

replace github.com/bsm/bps => ../

require (
	cloud.google.com/go/pubsub v1.4.0
	github.com/bsm/bps v0.0.0-20200605133303-c9abd8b1e2f6
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	google.golang.org/api v0.28.0
)
