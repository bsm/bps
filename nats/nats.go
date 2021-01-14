// Package nats implements nats.io backend adapter.
//
// Both bps.NewPublisher and bps.NewSubscriber support:
//
//   client_cert, client_key
//     nats client certificate and key file paths.
//
package nats

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/bsm/bps"
	"github.com/nats-io/nats.go"
)

func init() {
	bps.RegisterPublisher("nats", func(ctx context.Context, u *url.URL) (bps.Publisher, error) {
		natsURL, opts, err := prepareConnectionArgs(u)
		if err != nil {
			return nil, err
		}
		return NewPublisher(natsURL, opts...)
	})
	bps.RegisterSubscriber("nats", func(ctx context.Context, u *url.URL) (bps.Subscriber, error) {
		natsURL, opts, err := prepareConnectionArgs(u)
		if err != nil {
			return nil, err
		}
		return NewSubscriber(natsURL, opts...)
	})
}

// ----------------------------------------------------------------------------

type publisher struct {
	conn *nats.Conn
}

// NewPublisher inits nats.io-backed publisher.
func NewPublisher(natsURL string, opts ...nats.Option) (bps.Publisher, error) {
	conn, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, err
	}

	return &publisher{
		conn: conn,
	}, nil
}

func (p *publisher) Topic(name string) bps.PubTopic {
	return &pubTopic{
		conn: p.conn,
		name: name,
	}
}

func (p *publisher) Close() error {
	if err := p.conn.Flush(); err != nil {
		return err
	}

	p.conn.Close()
	return nil
}

// ----------------------------------------------------------------------------

type pubTopic struct {
	conn *nats.Conn
	name string
}

func (t *pubTopic) Publish(ctx context.Context, msg *bps.PubMessage) error {
	return t.conn.Publish(t.name, msg.Data)
}

// ----------------------------------------------------------------------------

type subscriber struct {
	conn *nats.Conn
}

// NewSubscriber inits nats.io-backed subscriber.
func NewSubscriber(natsURL string, opts ...nats.Option) (bps.Subscriber, error) {
	conn, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, err
	}

	return &subscriber{
		conn: conn,
	}, nil
}

func (s *subscriber) Topic(name string) bps.SubTopic {
	return &subTopic{
		conn: s.conn,
		name: name,
	}
}

func (s *subscriber) Close() error {
	s.conn.Close()
	return nil
}

// ----------------------------------------------------------------------------

type subTopic struct {
	conn *nats.Conn
	name string
}

func (t *subTopic) Subscribe(handler bps.Handler, options ...bps.SubOption) (bps.Subscription, error) {
	// options are handled only for checking - return error if user expects smth that is not supported by nats:
	opts := (&bps.SubOptions{
		StartAt: bps.PositionNewest,
	}).Apply(options)
	if opts.StartAt != bps.PositionNewest {
		return nil, fmt.Errorf("start position %s is not supported by this implementation (PositionNewest is the only option)", opts.StartAt)
	}
	// error handler is silently ignored, as it is supported conn-wide, not per-subscription.

	sub, err := t.conn.Subscribe(
		t.name,
		func(msg *nats.Msg) {
			handler.Handle(bps.RawSubMessage(msg.Data))
		},
	)
	if err != nil {
		return nil, err
	}
	return &subscription{
		sub: sub,
	}, nil
}

// ----------------------------------------------------------------------------

type subscription struct {
	sub *nats.Subscription
}

func (s *subscription) Close() error {
	return s.sub.Unsubscribe()
}

// ----------------------------------------------------------------------------

// prepareConnectionArgs parses args for NewSubscriber/NewPublisher from URL.
func prepareConnectionArgs(u *url.URL) (
	natsURL string,
	opts []nats.Option,
	err error,
) {
	q := u.Query()

	// converted/cleaned up version of a BPS URL:
	natsURL = (&url.URL{
		Scheme: "nats",
		Host:   u.Host, // host or host:port
	}).String()

	if clientCert, clientKey := q.Get("client_cert"), q.Get("client_key"); clientCert != "" || clientKey != "" {
		if clientCert == "" {
			return "", nil, errors.New("no client_cert provided")
		}
		if clientKey == "" {
			return "", nil, errors.New("no client_key provided")
		}
		opts = append(opts, nats.ClientCert(clientCert, clientKey))
	}

	return natsURL, opts, nil
}
