package stan_test

import "os"

const clusterID = "test-cluster" // clusterID holds (default) nats-streaming cluster ID: https://hub.docker.com/_/nats-streaming

var stanAddr string

func init() {
	stanAddr = "localhost:4222"
	if v := os.Getenv("STAN_ADDR"); v != "" {
		stanAddr = v
	}
}
