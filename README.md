# BPS

[![GoDoc](https://godoc.org/github.com/bsm/bps?status.svg)](https://godoc.org/github.com/bsm/bps)
[![Build Status](https://travis-ci.org/bsm/bps.svg?branch=master)](https://travis-ci.org/bsm/bps)
[![Go Report Card](https://goreportcard.com/badge/github.com/bsm/bps)](https://goreportcard.com/report/github.com/bsm/bps)

Multi-backend abstraction for message processing and pubsub queues.

## Documentation

For documentation and examples, please see https://godoc.org/github.com/bsm/bps.

## Install

```shell
go get -u github.com/bsm/bps
```

## Backends

* [Google PubSub](https://godoc.org/github.com/bsm/bps/google)
* [File](https://godoc.org/github.com/bsm/bps/file)
* [Kafka](https://godoc.org/github.com/bsm/bps/kafka)

## Publishing

```go
package main

import (
	"context"
	"fmt"

	"github.com/bsm/bps"
)

func main() {
	ctx := context.Background()
	pub := bps.NewInMemPublisher()
	defer pub.Close()

	topicA := pub.Topic("topic-a")
	topicB := pub.Topic("topic-b")

	topicA.Publish(ctx, &bps.Message{
		Data: []byte("message-1"),
	})
	topicB.Publish(ctx, &bps.Message{
		Data: []byte("message-2"),
	})
	topicA.Publish(ctx, &bps.Message{
		Data: []byte("message-2"),
	})

	fmt.Println(len(topicA.(*bps.InMemTopic).Messages()))
	fmt.Println(len(topicB.(*bps.InMemTopic).Messages()))

}
```
