module github.com/bsm/bps/pubsub

go 1.13

replace github.com/bsm/bps => ../

require (
	cloud.google.com/go v0.60.0 // indirect
	cloud.google.com/go/pubsub v1.4.0
	github.com/bsm/bps v0.0.0-20200605133303-c9abd8b1e2f6
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	go.opencensus.io v0.22.4 // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208 // indirect
	google.golang.org/api v0.28.0
	google.golang.org/grpc v1.30.0 // indirect
)
