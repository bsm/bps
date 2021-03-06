package bps

import (
	"context"
	"fmt"
	"net/url"
	"sync"
)

var (
	pubReg   = map[string]PublisherFactory{}
	pubRegMu sync.Mutex
)

// Publisher defines the main publisher interface.
type Publisher interface {
	// Topic returns a topic handle by name.
	Topic(name string) PubTopic
	// Close closes the producer connection.
	Close() error
}

// NewPublisher inits to a publisher via URL.
//
//   pub, err := bps.NewPublisher(context.TODO(), "kafka://10.0.0.1:9092,10.0.0.2:9092,10.0.0.3:9092/namespace")
func NewPublisher(ctx context.Context, urlStr string) (Publisher, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	pubRegMu.Lock()
	factory, ok := pubReg[u.Scheme]
	pubRegMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("unknown URL scheme %q", u.Scheme)
	}
	return factory(ctx, u)
}

// --------------------------------------------------------------------

// PublisherFactory constructs a publisher from a URL.
type PublisherFactory func(context.Context, *url.URL) (Publisher, error)

// RegisterPublisher registers a new protocol with a scheme and a corresponding
// PublisherFactory.
func RegisterPublisher(scheme string, factory PublisherFactory) {
	pubRegMu.Lock()
	defer pubRegMu.Unlock()

	if _, exists := pubReg[scheme]; exists {
		panic("protocol " + scheme + " already registered")
	}
	pubReg[scheme] = factory
}

// --------------------------------------------------------------------

// PubMessage represents a single message for publishing.
type PubMessage struct {
	// ID is an optional message identifier.
	// It may not be supported by some implementations (then it is ignored).
	// Or may be used just to calculate partition the message.
	ID string `json:"id,omitempty"`

	// Data is the message payload.
	Data []byte `json:"data,omitempty"`

	// Attributes contains optional key-value labels.
	// It may not be supported by some implementations (then it is ignored).
	Attributes map[string]string `json:"attributes,omitempty"`
}

// PubTopic is a publisher handle to a topic.
type PubTopic interface {
	// Publish publishes a message to the topic.
	Publish(context.Context, *PubMessage) error
}

// --------------------------------------------------------------------

// InMemPublisher is an in-memory publisher implementation which can be used for tests.
type InMemPublisher struct {
	topics map[string]*InMemPubTopic
	mu     sync.RWMutex
}

// NewInMemPublisher returns an initialised publisher.
func NewInMemPublisher() *InMemPublisher {
	return &InMemPublisher{
		topics: make(map[string]*InMemPubTopic),
	}
}

// Topic implements Publisher interface. It will auto-provision a topic if it does not exist.
func (p *InMemPublisher) Topic(name string) PubTopic {
	p.mu.RLock()
	topic, ok := p.topics[name]
	p.mu.RUnlock()

	if ok {
		return topic
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if topic, ok = p.topics[name]; !ok {
		topic = new(InMemPubTopic)
		p.topics[name] = topic
	}
	return topic
}

// Close implements Publisher.
func (*InMemPublisher) Close() error {
	return nil
}

// InMemPubTopic is an in-memory implementation of a Topic.
// Useful for tests.
type InMemPubTopic struct {
	messages []*PubMessage
	mu       sync.RWMutex
}

// Publish implements Topic.
func (t *InMemPubTopic) Publish(_ context.Context, msg *PubMessage) error {
	t.mu.Lock()
	t.messages = append(t.messages, msg)
	t.mu.Unlock()
	return nil
}

// Messages returns published messages.
func (t *InMemPubTopic) Messages() []*PubMessage {
	t.mu.RLock()
	messages := t.messages
	t.mu.RUnlock()
	return messages
}
