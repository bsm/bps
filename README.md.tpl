# BPS

[![Build Status](https://travis-ci.org/bsm/bps.svg?branch=master)](https://travis-ci.org/bsm/bps)

[![GoDoc](https://godoc.org/github.com/bsm/bps?status.svg)](https://pkg.go.dev/github.com/bsm/bps?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/bsm/bps)](https://goreportcard.com/report/github.com/bsm/bps)

[TODO: rubygems/rubydoc badges when published]

Multi-backend abstraction for message processing and pubsub queues for Go and Ruby.

## Documentation

Check auto-generated documentation:

- go: [GoDoc](https://pkg.go.dev/github.com/bsm/bps)
- ruby: TODO (rubydoc link(s) once gem published)

## Install

```shell
# go:
go get -u github.com/bsm/bps

# ruby:
bundle add 'bps-kafka'
```

## Backends: Go

- [Google PubSub](https://godoc.org/github.com/bsm/bps/pubsub)
- [File](https://godoc.org/github.com/bsm/bps/file)
- [Kafka](https://godoc.org/github.com/bsm/bps/kafka)

## Backends: Ruby

- TODO: bps-kafka rubydoc link

## Publishing: Go

```go
package main

import (
	"context"
	"fmt"

	"github.com/bsm/bps"
)

func main() {{ "ExamplePublisher" | code }}
```

## Publishing: Ruby

```ruby
# TODO
```

## Subscribing: Go

```go
package main

import (
	"context"
	"fmt"

	"github.com/bsm/bps"
)

func main() {{ "ExamplePublisher" | code }}
```

### Subscribing: Ruby

```ruby
# TODO
```
