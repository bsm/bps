# BPS

[![Build Status](https://travis-ci.org/bsm/bps.svg?branch=master)](https://travis-ci.org/bsm/bps)
[![GoDoc](https://godoc.org/github.com/bsm/bps?status.svg)](https://pkg.go.dev/github.com/bsm/bps?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/bsm/bps)](https://goreportcard.com/report/github.com/bsm/bps)
[![Gem Version](https://badge.fury.io/rb/bps.svg)](https://badge.fury.io/rb/bps)

Multi-backend abstraction for message processing and pubsub queues for Go and Ruby.

## Documentation

Check auto-generated documentation:

- go: [GoDoc](https://pkg.go.dev/github.com/bsm/bps)
- ruby: [RubyDoc](https://www.rubydoc.info/gems/bps)

## Install

```shell
# go:
go get -u github.com/bsm/bps
go get -u github.com/bsm/bps/kafka
go get -u github.com/bsm/bps/nats
go get -u github.com/bsm/bps/stan

# ruby:
bundle add 'bps-kafka'
bundle add 'bps-nats'
bundle add 'bps-stan'
```

## Backends: Go

- [Google PubSub](https://godoc.org/github.com/bsm/bps/pubsub)
- [File](https://godoc.org/github.com/bsm/bps/file)
- [Kafka](https://godoc.org/github.com/bsm/bps/kafka)
- [NATS](https://godoc.org/github.com/bsm/bps/nats)
- [STAN (NATS-streaming)](https://godoc.org/github.com/bsm/bps/stan)

## Backends: Ruby

- [Kafka](https://rubygems.org/gems/bps-kafka)
- [NATS](https://rubygems.org/gems/bps-nats)
- [STAN (NATS-streaming)](https://rubygems.org/gems/bps-stan)

## Publishing: Go

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

	_ = topicA.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-1"),
	})
	_ = topicB.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-2"),
	})
	_ = topicA.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-2"),
	})

	fmt.Println(len(topicA.(*bps.InMemPubTopic).Messages()))
	fmt.Println(len(topicB.(*bps.InMemPubTopic).Messages()))

}
```

## Publishing: Ruby

```ruby
require 'bps/kafka'

pub = BPS::Publisher.resolve('kafka://localhost:9092')
top = pub.topic('topic')

top.publish('foo')
top.publish('bar')

pub.close
```

To seed multiple brokers, use:

```ruby
BPS::Publisher.resolve('kafka://10.0.0.1,10.0.0.2,10.0.0.3:9092')
```

If your brokers are on different ports, try:

```ruby
BPS::Publisher.resolve('kafka://10.0.0.1%3A9092,10.0.0.2%3A9093,10.0.0.3%3A9094')
```

## Subscribing: Go

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

	_ = topicA.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-1"),
	})
	_ = topicB.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-2"),
	})
	_ = topicA.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-2"),
	})

	fmt.Println(len(topicA.(*bps.InMemPubTopic).Messages()))
	fmt.Println(len(topicB.(*bps.InMemPubTopic).Messages()))

}
```
