module github.com/bsm/bps/pubsub

go 1.13

replace github.com/bsm/bps => ../

require (
	cloud.google.com/go/pubsub v1.0.1
	cloud.google.com/go/storage v1.0.0 // indirect
	github.com/bsm/bps v0.0.0-20190924074548-720a29804376
	github.com/golang/groupcache v0.0.0-20191002201903-404acd9df4cc // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.0
	go.opencensus.io v0.22.1 // indirect
	go.uber.org/multierr v1.5.0
	golang.org/x/exp v0.0.0-20191002040644-a1355ae1e2c3 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	google.golang.org/api v0.11.0
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191009194640-548a555dbc03 // indirect
	google.golang.org/grpc v1.24.0 // indirect
)
