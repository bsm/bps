module github.com/bsm/bps/pubsub

go 1.16

replace github.com/bsm/bps => ../

require (
	cloud.google.com/go/pubsub v1.17.1
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/bsm/ginkgo v1.16.5
	github.com/bsm/gomega v1.17.0
	google.golang.org/api v0.64.0
)
