// Package pubsub provides a Google PubSub abstraction.
package pubsub

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	native "cloud.google.com/go/pubsub"
	"github.com/bsm/bps"
	"go.uber.org/multierr"
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
	bps.RegisterSubscriber("pubsub", func(ctx context.Context, u *url.URL) (bps.Subscriber, error) {
		projectID := u.Host
		query := u.Query()
		if v := query.Get("project_id"); v != "" {
			projectID = v
		}
		return NewSubscriber(ctx, projectID)
	})
}

// --------------------------------------------------------------------

// Publisher wraps a google pubsub client and implements the bps.Publisher interface.
type Publisher struct {
	client   *native.Client
	settings *native.PublishSettings
	topics   map[string]*PubTopic
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
		topics:   make(map[string]*PubTopic),
	}, nil
}

// Topic implements the bps.Publisher interface.
func (p *Publisher) Topic(name string) bps.PubTopic {
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
		topic = &PubTopic{topic: nt}
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

// PubTopic wraps a pubsub topic.
type PubTopic struct {
	topic *native.Topic
	last  atomic.Value
}

// Publish implements the bps.Topic interface.
func (t *PubTopic) Publish(ctx context.Context, msg *bps.PubMessage) error {
	res := t.topic.Publish(ctx, &native.Message{
		ID:         msg.ID,
		Data:       msg.Data,
		Attributes: msg.Attributes,
	})
	t.last.Store(res)
	return nil
}

// PublishBatch implements the bps.Topic interface.
func (t *PubTopic) PublishBatch(ctx context.Context, batch []*bps.PubMessage) error {
	for _, msg := range batch {
		if err := t.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Topic returns the native pubsub Topic. Use at your own risk!
func (t *PubTopic) Topic() *native.Topic {
	return t.topic
}

func (t *PubTopic) wait(ctx context.Context) error {
	if v := t.last.Load(); v != nil {
		_, err := v.(*native.PublishResult).Get(ctx)
		return err
	}
	return nil
}

// --------------------------------------------------------------------

// Subscriber is a Google PubSub wrapper that implements bps.Subscriber interface.
type Subscriber struct {
	client *native.Client
}

// NewSubscriber inits a subscriber.
// It starts handling from the newest available message (published after subscribing).
// Google PubSub may re-deliver successfully handled messages.
func NewSubscriber(ctx context.Context, projectID string) (*Subscriber, error) {
	client, err := native.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		client: client,
	}, nil
}

// Topic returns a subcriber topic handle.
func (s *Subscriber) Topic(name string) bps.SubTopic {
	return &subTopic{
		client: s.client,
		name:   name,
	}
}

// Close closes the client.
func (s *Subscriber) Close() error {
	return s.client.Close()
}

// --------------------------------------------------------------------

type subTopic struct {
	client *native.Client
	name   string
}

func (t *subTopic) Subscribe(ctx context.Context, handler bps.Handler, _ ...bps.SubOption) error {
	// usual way to "unsubscribe" from Google PubSub is to cancel context:
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sub, err := t.client.CreateSubscription(ctx, bps.GenClientID(), native.SubscriptionConfig{
		Topic: t.client.Topic(t.name),
	})
	if err != nil {
		return err
	}
	defer func() {
		// give subscription 5s to delete:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = sub.Delete(ctx)
	}()

	// sub.Receive calls f concurrently from multiple goroutines
	psHandler := &pubsubHandler{Handler: handler}
	handler = psHandler // reassign original handler not to call it by mistake

	err = sub.Receive(ctx, func(ctx context.Context, msg *native.Message) {
		defer msg.Nack() // only first call to Ack/Nack matters, so it's safe

		if err := handler.Handle(bps.RawSubMessage(msg.Data)); err != nil {
			cancel() // unsubscribe on any error
			if !errors.Is(err, bps.Done) {
				return // do not proceed with Ack-ing
			}
		}
		msg.Ack() // no error returned, msg will be re-delivered on Ack failure
	})

	// TODO: may need to suppress sub.Receive's err:
	//       pubsub native lib is based on streaming pull, which is expected to terminate with error:
	//       https://cloud.google.com/pubsub/docs/pull#streamingpull_has_a_100_error_rate_this_is_to_be_expected
	//       StreamingPull streams are always terminated with a non-OK status.
	//       Note that, unlike in regular RPCs, the status here is simply an indication that the stream has been broken, not that requests are failing.
	//       Therefore, while the StreamingPull API may have a seemingly surprising 100% error rate, this is by design.

	return multierr.Combine(err, psHandler.LastErr()) // nil errs are ignored my multierr
}

// ----------------------------------------------------------------------------

// pubsubHandler is a thread-safe handler, that captures non-nil/non-bps.Done errors.
type pubsubHandler struct {
	mu      sync.Mutex
	lastErr error // last non-nil/non-bps.Done error

	Handler bps.Handler
}

func (h *pubsubHandler) Handle(msg bps.SubMessage) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.Handler.Handle(msg)
	if err != nil && !errors.Is(err, bps.Done) {
		h.lastErr = err
	}
	return err
}

func (h *pubsubHandler) LastErr() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.lastErr
}
