// Package nats abstracts publish-subscribe https://nats.io backend.
package nats

import (
	"context"
	"errors"
	"fmt"

	"github.com/bsm/bps"
	"github.com/nats-io/stan.go"
	"go.uber.org/multierr"
)

type publisher struct {
	conn stan.Conn
}

// NewPublisher constructs a new nats.io-backed publisher.
func NewPublisher(stanClusterID, clientID string, options ...stan.Option) (bps.Publisher, error) {
	conn, err := stan.Connect(stanClusterID, clientID, options...)
	if err != nil {
		return nil, err
	}
	return &publisher{conn: conn}, nil
}

func (p *publisher) Topic(name string) bps.Topic {
	return &topic{
		conn:    p.conn,
		subject: name,
	}
}

func (p *publisher) Close() error {
	return p.conn.Close()
}

type topic struct {
	conn    stan.Conn
	subject string
}

func (t *topic) Publish(_ context.Context, msg *bps.PubMessage) error {
	return t.conn.Publish(t.subject, msg.Data)
}

func (t *topic) PublishBatch(ctx context.Context, messages []*bps.PubMessage) error {
	for i, msg := range messages {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("publish message %d of %d: %w", i, len(messages), err)
		}
		if err := t.Publish(ctx, msg); err != nil {
			return fmt.Errorf("publish message %d of %d: %w", i, len(messages), err)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

type subscriber struct {
	conn stan.Conn
}

// NewSubscriber constructs a new nats.io-backed subscriber.
func NewSubscriber(stanClusterID, clientID string, options ...stan.Option) (bps.Subscriber, error) {
	conn, err := stan.Connect(stanClusterID, clientID, options...)
	if err != nil {
		return nil, err
	}
	return &subscriber{conn: conn}, nil
}

func (s *subscriber) Subscribe(ctx context.Context, topic string, handler bps.Handler) error {
	var (
		sub        stan.Subscription
		err        error
		handlerErr error
	)
	sub, err = s.conn.Subscribe(topic, func(msg *stan.Msg) {
		defer sub.Unsubscribe()

		if err := handler.Handle(bps.RawSubMessage(msg.Data)); errors.Is(err, bps.Done) {
			return
		} else if err != nil {
			handlerErr = err
			return
		}
	})
	return multierr.Combine(err, handlerErr)
}

func (s *subscriber) Close() error {
	return s.conn.Close()
}
