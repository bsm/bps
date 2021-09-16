package stan_test

import "os"

const clusterID = "test-cluster" // clusterID holds (default) nats-streaming cluster ID: https://hub.docker.com/_/nats-streaming

var stanAddrs string

func init() {
	stanAddrs = "localhost:4222"
	if v := os.Getenv("STAN_ADDRS"); v != "" {
		stanAddrs = v
	}
}
