module github.com/bsm/bps/pubsub

go 1.13

replace github.com/bsm/bps => ../

require (
	cloud.google.com/go/pubsub v1.0.1
	cloud.google.com/go/storage v1.0.0 // indirect
	github.com/bsm/bps v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/exp v0.0.0-20190919035709-81c71964d733 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/tools v0.0.0-20190920225731-5eefd052ad72 // indirect
	google.golang.org/api v0.10.0
	google.golang.org/appengine v1.6.3 // indirect
	google.golang.org/genproto v0.0.0-20190916214212-f660b8655731 // indirect
	google.golang.org/grpc v1.23.1 // indirect
)
