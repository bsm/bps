package bps

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
)

var (
	subReg   = map[string]SubscriberFactory{}
	subRegMu sync.Mutex
)

// ----------------------------------------------------------------------------

type done string

func (s done) Error() string { return string(s) }

// Done is intended to communicate "stop operation" from callbacks.
const Done = done("done")

// ----------------------------------------------------------------------------

// SubMessage defines a subscription message details.
type SubMessage interface {
	// Data returns raw (serialized) message data.
	Data() []byte
}

// RawSubMessage is an adapter for raw slice of bytes that behaves as a SubMessage.
type RawSubMessage []byte

// Data returns raw message bytes.
func (m RawSubMessage) Data() []byte {
	return m
}

// ----------------------------------------------------------------------------

// Handler defines a message handler.
// Consuming can be stopped by returning bps.Done.
type Handler interface {
	Handle(SubMessage) error
}

// HandlerFunc is a func-based handler adapter.
type HandlerFunc func(SubMessage) error

// Handle handles a single message.
func (f HandlerFunc) Handle(msg SubMessage) error {
	return f(msg)
}

// ----------------------------------------------------------------------------

// SubStart defines starting position to consume messages.
type SubStart int

// SubStart options.
const (
	// Newest tells to start consuming messages from the newest available
	// (published AFTER subscribing).
	Newest SubStart = iota

	// Oldest tells to start consuming messages from the oldest available
	// (published BEFORE subscribing).
	Oldest
)

// SubOptions holds subscription options.
type SubOptions struct {
	// Start defines starting position to consume messages.
	// May not be supported by some implementations.
	// Default: implementation-specific (Newest is recommended).
	Start SubStart
}

// Apply configures SubOptions struct by applying each single SubOption one by one.
//
// It is meant to be used by pubsub implementations like this:
//
//   func (s *SubImpl) Subscribe(..., options ...bps.SubOption) error {
//     opts := (&bps.SubOptions{
//       // implementation-specific defaults
//     }).Apply(options)
//     ...
//   }
//
func (o *SubOptions) Apply(options []SubOption) *SubOptions {
	if o == nil {
		o = new(SubOptions)
	}
	for _, opt := range options {
		opt(o)
	}
	return o
}

// SubOption defines a single subscription option.
type SubOption func(*SubOptions)

// Start configures subscription start position.
func Start(start SubStart) SubOption {
	return func(o *SubOptions) {
		o.Start = start
	}
}

// ----------------------------------------------------------------------------

// Subscriber defines the main subscriber interface.
type Subscriber interface {
	// Subscribe subscribes for topic messages and blocks till context is cancelled or error occurs or bps.Done is returned.
	Subscribe(ctx context.Context, topic string, handler Handler, opts ...SubOption) error
	// Close closes the subscriber connection.
	Close() error
}

// NewSubscriber inits to a subscriber via URL.
//
//   sub, err := bps.NewSubscriber(context.TODO(), "kafka://10.0.0.1:9092,10.0.0.2:9092,10.0.0.3:9092/namespace")
//
func NewSubscriber(ctx context.Context, urlStr string) (Subscriber, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	subRegMu.Lock()
	factory, ok := subReg[u.Scheme]
	subRegMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("unknown URL scheme %q", u.Scheme)
	}
	return factory(ctx, u)
}

// SubscriberFactory constructs a subscriber from a URL.
type SubscriberFactory func(context.Context, *url.URL) (Subscriber, error)

// RegisterSubscriber registers a new protocol with a scheme and a corresponding
// SubscriberFactory.
func RegisterSubscriber(scheme string, factory SubscriberFactory) {
	subRegMu.Lock()
	defer subRegMu.Unlock()

	if _, exists := subReg[scheme]; exists {
		panic("protocol " + scheme + " already registered")
	}
	subReg[scheme] = factory
}

// ----------------------------------------------------------------------------

// InMemSubscriber is a subscriber, that consumes messages from seeded data.
// It is useful mainly for testing.
type InMemSubscriber struct {
	mu   sync.Mutex
	msgs map[string][]SubMessage
}

// NewInMemSubscriber returns new subscriber, that consumes messages from seeded data.
func NewInMemSubscriber(messagesByTopic map[string][]SubMessage) *InMemSubscriber {
	byTopic := make(map[string][]SubMessage, len(messagesByTopic))
	for topic, msgs := range messagesByTopic {
		byTopic[topic] = msgs
	}
	return &InMemSubscriber{
		msgs: byTopic,
	}
}

// Subscribe subscribes to in-memory messages by topic.
// It starts handling from the first (oldest) available message.
func (s *InMemSubscriber) Subscribe(ctx context.Context, topic string, handler Handler, _ ...SubOption) error {
	for {
		// check ctx/cancel BEFORE shift-ing each message:
		if err := ctx.Err(); err != nil {
			return err
		}

		msg, ok := s.shiftMessage(topic)
		if !ok {
			return nil
		}

		if err := handler.Handle(msg); errors.Is(err, Done) {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func (s *InMemSubscriber) shiftMessage(topic string) (SubMessage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msgs := s.msgs[topic]

	if len(msgs) == 0 {
		return nil, false
	}

	s.msgs[topic] = msgs[1:]
	return msgs[0], true
}

// Close forgets any pending messages.
func (s *InMemSubscriber) Close() error {
	s.mu.Lock()
	s.msgs = nil
	s.mu.Unlock()
	return nil
}
