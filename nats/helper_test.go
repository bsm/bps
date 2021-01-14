package nats_test

import "os"

var natsAddr string

func init() {
	natsAddr = "localhost:4222"
	if v := os.Getenv("NATS_ADDR"); v != "" {
		natsAddr = v
	}
}
