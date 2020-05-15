// Package pubsub provides a Google PubSub abstraction.
package pubsub

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	native "cloud.google.com/go/pubsub"
	"github.com/bsm/bps"
	"google.golang.org/api/option"
)

func init() {
	bps.RegisterPublisher("pubsub", func(ctx context.Context, u *url.URL) (bps.Publisher, error) {
		projectID := u.Host
		query := u.Query()
		if v := query.Get("project_id"); v != "" {
			projectID = v
		}
		return NewPublisher(ctx, projectID, nil)
	})
}

// --------------------------------------------------------------------

// Publisher wraps a google pubsub client and implements the bps.Publisher interface.
type Publisher struct {
	client   *native.Client
	settings *native.PublishSettings
	topics   map[string]*Topic
	mu       sync.RWMutex
}

// NewPublisher inits a publisher.
func NewPublisher(ctx context.Context, projectID string, settings *native.PublishSettings, opts ...option.ClientOption) (*Publisher, error) {
	client, err := native.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		client:   client,
		settings: settings,
		topics:   make(map[string]*Topic),
	}, nil
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
		nt := p.client.Topic(name)
		if p.settings != nil {
			nt.PublishSettings = *p.settings
		}
		topic = &Topic{topic: nt}
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
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	var err error
	for name, t := range p.topics {
		if e := t.wait(ctx); e != nil {
			err = e
		}
		delete(p.topics, name)
	}
	return err
}

// Client exposes the native client. Use at your own risk!
func (p *Publisher) Client() *native.Client {
	return p.client
}

// --------------------------------------------------------------------

// Topic wraps a pubsub topic.
type Topic struct {
	topic *native.Topic
	last  atomic.Value
}

// Publish implements the bps.Topic interface.
func (t *Topic) Publish(ctx context.Context, msg *bps.PubMessage) error {
	res := t.topic.Publish(ctx, &native.Message{
		ID:         msg.ID,
		Data:       msg.Data,
		Attributes: msg.Attributes,
	})
	t.last.Store(res)
	return nil
}

// PublishBatch implements the bps.Topic interface.
func (t *Topic) PublishBatch(ctx context.Context, batch []*bps.PubMessage) error {
	for _, msg := range batch {
		if err := t.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Topic returns the native pubsub Topic. Use at your own risk!
func (t *Topic) Topic() *native.Topic {
	return t.topic
}

func (t *Topic) wait(ctx context.Context) error {
	if v := t.last.Load(); v != nil {
		_, err := v.(*native.PublishResult).Get(ctx)
		return err
	}
	return nil
}
