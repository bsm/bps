// Package nats abstracts publish-subscribe https://nats.io backend.
package nats

import (
	"context"

	"github.com/bsm/bps"
	"github.com/nats-io/stan.go"
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
	panic("TODO")
}

func (p *publisher) Close() error {
	return p.conn.Close()
}

type topic struct {
}

func (t *topic) Publish(context.Context, *bps.PubMessage) error {
	panic("TODO")
}

func (t *topic) PublishBatch(context.Context, []*bps.PubMessage) error {
	panic("TODO")
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
	panic("TODO")
}

func (s *subscriber) Close() error {
	return s.conn.Close()
}
