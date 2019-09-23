// Package google provides a Google PubSub abstraction.
package google

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
	"github.com/bsm/bps"
	"google.golang.org/api/option"
)

func init() {
	bps.RegisterPublisher("google", func(ctx context.Context, u *url.URL) (bps.Publisher, error) {
		projectID := u.Host
		query := u.Query()
		if v := query.Get("project_id"); v != "" {
			projectID = v
		}
		return NewPublisher(ctx, projectID)
	})
}

// --------------------------------------------------------------------

// Publisher wraps a google pubsub client and implements the bps.Publisher interface.
type Publisher struct {
	client *pubsub.Client

	topics map[string]*Topic
	mu     sync.RWMutex
}

// NewPublisher inits a publisher.
func NewPublisher(ctx context.Context, projectID string, opts ...option.ClientOption) (*Publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}
	return &Publisher{client: client, topics: make(map[string]*Topic)}, nil
}

// Topic implements the bps.Publisher interface.
func (p *Publisher) Topic(name string) bps.Topic {
	p.mu.RLock()
	topic, ok := p.topics[name]
	p.mu.RUnlock()

	if ok {
		return topic
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if topic, ok = p.topics[name]; !ok {
		topic = &Topic{topic: p.client.Topic(name)}
		p.topics[name] = topic
	}
	return topic
}

// Close implements the bps.Publisher interface.
func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// stop all
	for _, t := range p.topics {
		t.topic.Stop()
	}

	// wait for all
	for name, t := range p.topics {
		t.wait()
		delete(p.topics, name)
	}
	return nil
}

// Client exposes the native client. Use at your own risk!
func (p *Publisher) Client() *pubsub.Client {
	return p.client
}

// --------------------------------------------------------------------

// Topic wraps a pubsub topic.
type Topic struct {
	topic *pubsub.Topic
	last  atomic.Value
}

// Publish implements the bps.Topic interface.
func (t *Topic) Publish(ctx context.Context, msg *bps.Message) error {
	res := t.topic.Publish(ctx, &pubsub.Message{
		ID:         msg.ID,
		Data:       msg.Data,
		Attributes: msg.Attributes,
	})
	t.last.Store(res)
	return nil
}

// PublishBatch implements the bps.Topic interface.
func (t *Topic) PublishBatch(ctx context.Context, batch []*bps.Message) error {
	for _, msg := range batch {
		if err := t.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Topic returns the native pubsub Topic. Use at your own risk!
func (t *Topic) Topic() *pubsub.Topic {
	return t.topic
}

func (t *Topic) wait() {
	if v := t.last.Load(); v != nil {
		<-v.(*pubsub.PublishResult).Ready()
	}
}
