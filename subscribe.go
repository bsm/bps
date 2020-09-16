package bps

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/bsm/bps/internal/concurrent"
)

var (
	subReg   = map[string]SubscriberFactory{}
	subRegMu sync.Mutex
)

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
	Handle(SubMessage)
}

// HandlerFunc is a func-based handler adapter.
type HandlerFunc func(SubMessage)

// Handle handles a single message.
func (f HandlerFunc) Handle(msg SubMessage) {
	f(msg)
}

// SafeHandler wraps a handler with a mutex to synchronize access.
// It is intended to be used only by subscriber implementations which need it.
// It shouldn't be used by lib consumer.
func SafeHandler(h Handler) Handler {
	if _, ok := h.(*safeHandler); ok {
		return h
	}
	return &safeHandler{Handler: h}
}

type safeHandler struct {
	Handler
	mu sync.Mutex
}

// Handle synchronizes underlying Handler.Handle calls.
func (h *safeHandler) Handle(msg SubMessage) {
	h.mu.Lock()
	h.Handler.Handle(msg)
	h.mu.Unlock()
}

// ----------------------------------------------------------------------------

// StartPosition defines starting position to consume messages.
type StartPosition string

// StartPosition options.
const (
	// PositionNewest tells to start consuming messages from the newest available
	// (published AFTER subscribing).
	PositionNewest StartPosition = "newest"

	// PositionOldest tells to start consuming messages from the oldest available
	// (published BEFORE subscribing).
	PositionOldest StartPosition = "oldest"
)

// SubOptions holds subscription options.
type SubOptions struct {
	// StartAt defines starting position to consume messages.
	// May not be supported by some implementations.
	// Default: implementation-specific (PositionNewest is recommended).
	StartAt StartPosition
	// ErrorHandler is a subscription error handler (system/implementation-specific errors).
	// Default: log errors to STDERR.
	ErrorHandler func(error)
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

	// apply some defaults:
	if o.ErrorHandler == nil {
		log := log.New(os.Stderr, "[bps] ", log.LstdFlags) // TODO: maybe make global?..
		o.ErrorHandler = func(err error) { log.Println(err) }
	}

	for _, opt := range options {
		opt(o)
	}
	return o
}

// SubOption defines a single subscription option.
type SubOption func(*SubOptions)

// StartAt configures subscription start position.
func StartAt(pos StartPosition) SubOption {
	return func(o *SubOptions) {
		o.StartAt = pos
	}
}

// WithErrorHandler configures subscription error handler.
func WithErrorHandler(h func(error)) SubOption {
	return func(o *SubOptions) {
		o.ErrorHandler = h
	}
}

// IgnoreSubscriptionErrors configures subscription to silently ignore errors.
func IgnoreSubscriptionErrors() SubOption {
	return func(o *SubOptions) {
		o.ErrorHandler = func(error) {} // noop
	}
}

// ----------------------------------------------------------------------------

// Subscription defines a subscription-manager interface.
type Subscription interface {
	// Close stops message handling and frees resources.
	// It is safe to be called multiple times.
	Close() error
}

// SubTopic defines a subscriber topic handle.
type SubTopic interface {
	// Subscribe subscribes for topic messages and handles them in background
	// till error occurs or bps.Done is returned.
	// Handler is guaranteed to be called synchronously (messages are handled one by one).
	Subscribe(handler Handler, opts ...SubOption) (Subscription, error)
}

// Subscriber defines the main subscriber interface.
type Subscriber interface {
	// Topic returns a subscriber topic handle.
	Topic(name string) SubTopic
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
	return &InMemSubscriber{msgs: byTopic}
}

// Topic returns named topic handle.
// Seeded messages are copied for each topic handle
func (s *InMemSubscriber) Topic(topic string) SubTopic {
	s.mu.Lock()
	defer s.mu.Unlock()

	return NewInMemSubTopic(s.msgs[topic])
}

// Close forgets seeded messages.
func (s *InMemSubscriber) Close() error {
	s.mu.Lock()
	s.msgs = nil
	s.mu.Unlock()
	return nil
}

// InMemSubTopic is a subscriber topic handle, that consumes messages from seeded data.
// It is useful mainly for testing.
type InMemSubTopic struct {
	mu   sync.Mutex
	msgs []SubMessage
}

// NewInMemSubTopic returns new seeded in-memory subscriber topic handle.
func NewInMemSubTopic(msgs []SubMessage) *InMemSubTopic {
	return &InMemSubTopic{msgs: msgs}
}

// Subscribe subscribes to in-memory messages by topic.
// It starts handling from the first (oldest) available message.
func (s *InMemSubTopic) Subscribe(handler Handler, _ ...SubOption) (Subscription, error) {
	sub := concurrent.NewGroup(context.Background())

	sub.Go(func() {
		for {
			select {
			case <-sub.Done():
				return
			default:
			}

			msg, ok := s.shiftMessage()
			if !ok {
				// no more messages, OK to return now
				return
			}

			handler.Handle(msg)
		}
	})

	return sub, nil
}

func (s *InMemSubTopic) shiftMessage() (SubMessage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msgs := s.msgs
	if len(msgs) == 0 {
		return nil, false
	}

	s.msgs = msgs[1:]
	return msgs[0], true
}
