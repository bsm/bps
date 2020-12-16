package nats_test

import "os"

const clusterID = "test-cluster" // clusterID holds (default) nats-streaming cluster ID: https://hub.docker.com/_/nats-streaming

var natsAddr string

func init() {
	natsAddr = "localhost:4222"
	if v := os.Getenv("NATS_ADDR"); v != "" {
		natsAddr = v
	}
}
