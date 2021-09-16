package nats_test

import "os"

var natsAddrs string

func init() {
	natsAddrs = "localhost:4222"
	if v := os.Getenv("NATS_ADDRS"); v != "" {
		natsAddrs = v
	}
}
