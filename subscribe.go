package bps

import (
	"context"
	"fmt"
	"net/url"
	"sync"
)

var (
	subReg   = map[string]SubscriberFactory{}
	subRegMu sync.Mutex
)

// ----------------------------------------------------------------------------

// SubMessage defines a subscription message details.
type SubMessage interface {
	Data() []byte
}

// Handler defines a message handler.
type Handler interface {
	Handle(context.Context, SubMessage) error
}

// HandlerFunc is a func-based handler adapter.
type HandlerFunc func(context.Context, SubMessage) error

// Handle handles a single message.
func (f HandlerFunc) Handle(ctx context.Context, msg SubMessage) error {
	return f(ctx, msg)
}

// ----------------------------------------------------------------------------

// SubscribeOptions defines subscribe options.
// Note: not every option may be supported by the underlying implementation.
type SubscribeOptions struct {
}

// ----------------------------------------------------------------------------

// Subscriber defines the main subscriber interface.
type Subscriber interface {
	// Subscribe subscribes for topic messages and blocks till context is cancelled or error occurs.
	//
	// Handler is provided an atomic message batch (entire batch either succeeds or fails).
	// Batch is guaranteed to contain at least 1 message.
	Subscribe(
		ctx context.Context,
		topic string,
		handler Handler,
		opts *SubscribeOptions,
	) error
	// Close closes the subscriber connection.
	Close() error
}

// NewSubscriber inits to a subscriber via URL.
//
//   sub, err := bps.NewSubscriber(context.TODO(), "kafka://10.0.0.1:9092,10.0.0.2:9092,10.0.0.3:9092/namespace")
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
