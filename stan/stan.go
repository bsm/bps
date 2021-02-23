// Package stan abstracts publish-subscribe https://docs.nats.io/developing-with-nats-streaming/streaming backend.
//
// WARNING/TODO: messages can be considered not acknowledged,
//               because DurableName for subscriptions is not used yet:
//               https://godoc.org/github.com/nats-io/stan.go#DurableName
//
// Both bps.NewPublisher and bps.NewSubscriber support:
//
//   client_id
//     nats-streaming client ID, [0-9A-Za-z_-] only.
//   client_cert, client_key
//     nats client certificate and key file paths.
//
// bps.NewSubscriber supports:
//
//   queue_group
//     optional queue group name for queue subscriptions: https://docs.nats.io/developing-with-nats/receiving/queues
//
package stan

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/bsm/bps"
	natsio "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

func init() {
	bps.RegisterPublisher("stan", func(ctx context.Context, u *url.URL) (bps.Publisher, error) {
		natsConn, clusterID, clientID, opts, err := prepareConnectionArgs(u)
		if err != nil {
			return nil, err
		}
		return newPublisherWithNatsConn(natsConn, clusterID, clientID, opts)
	})
	bps.RegisterSubscriber("stan", func(ctx context.Context, u *url.URL) (bps.Subscriber, error) {
		natsConn, clusterID, clientID, opts, err := prepareConnectionArgs(u)
		if err != nil {
			return nil, err
		}
		queueGroup := u.Query().Get("queue_group")
		return newSubscriberWithNatsConn(natsConn, clusterID, clientID, queueGroup, opts)
	})
}

type publisher struct {
	conn     stan.Conn
	natsConn *natsio.Conn // managed NATS connection, if any
}

// NewPublisher constructs a new STAN-backed publisher.
func NewPublisher(stanClusterID, clientID string, opts []stan.Option) (bps.Publisher, error) {
	return newPublisherWithNatsConn(nil, stanClusterID, clientID, opts)
}

func newPublisherWithNatsConn(natsConn *natsio.Conn, stanClusterID, clientID string, opts []stan.Option) (bps.Publisher, error) {
	c, err := stan.Connect(stanClusterID, clientID, opts...)
	if err != nil {
		return nil, err
	}

	return &publisher{
		conn:     c,
		natsConn: natsConn,
	}, nil
}

func (p *publisher) Topic(name string) bps.PubTopic {
	return &pubTopic{
		conn: p.conn,
		name: name,
	}
}

func (p *publisher) Close() error {
	err := p.conn.Close()

	// close managed nats connection, if any:
	if nc := p.natsConn; nc != nil {
		nc.Close()
	}

	return err
}

// ----------------------------------------------------------------------------

type pubTopic struct {
	conn stan.Conn
	name string
}

func (t *pubTopic) Publish(ctx context.Context, msg *bps.PubMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return t.conn.Publish(t.name, msg.Data)
}

// ----------------------------------------------------------------------------

type subscriber struct {
	conn       stan.Conn
	natsConn   *natsio.Conn // managed NATS connection, if any
	queueGroup string       // optional, switches between .Subscribe and .QueueSubscribe
}

// NewSubscriber constructs a new STAN-backed subscriber.
// By default, it starts handling from the newest available message (published after subscribing).
// If queueGroup is specified, all subscriptions will be queue ones: https://docs.nats.io/developing-with-nats/receiving/queues
func NewSubscriber(stanClusterID, clientID, queueGroup string, opts []stan.Option) (bps.Subscriber, error) {
	return newSubscriberWithNatsConn(nil, stanClusterID, clientID, queueGroup, opts)
}

func newSubscriberWithNatsConn(natsConn *natsio.Conn, stanClusterID, clientID, queueGroup string, opts []stan.Option) (bps.Subscriber, error) {
	c, err := stan.Connect(stanClusterID, clientID, opts...)
	if err != nil {
		return nil, err
	}

	return &subscriber{
		conn:       c,
		natsConn:   natsConn,
		queueGroup: queueGroup,
	}, nil
}

func (s *subscriber) Topic(name string) bps.SubTopic {
	return &subTopic{
		stan:       s.conn,
		name:       name,
		queueGroup: s.queueGroup,
	}
}

func (s *subscriber) Close() error {
	err := s.conn.Close()

	// close managed nats connection, if any:
	if nc := s.natsConn; nc != nil {
		nc.Close()
	}

	return err
}

// ----------------------------------------------------------------------------

type subTopic struct {
	stan       stan.Conn
	name       string
	queueGroup string
}

func (t *subTopic) Subscribe(handler bps.Handler, options ...bps.SubOption) (bps.Subscription, error) {
	opts := (&bps.SubOptions{
		StartAt: bps.PositionNewest,
	}).Apply(options)

	var startPos pb.StartPosition
	switch opts.StartAt {
	case bps.PositionNewest:
		startPos = pb.StartPosition_NewOnly
	case bps.PositionOldest:
		startPos = pb.StartPosition_First
	default:
		return nil, fmt.Errorf("start position %s is not supported by this implementation", opts.StartAt)
	}

	stanHandler := func(msg *stan.Msg) {
		handler.Handle(bps.RawSubMessage(msg.Data))
	}

	var (
		sub stan.Subscription
		err error
	)
	if t.queueGroup == "" {
		sub, err = t.stan.Subscribe(t.name, stanHandler, stan.StartAt(startPos))
	} else {
		sub, err = t.stan.QueueSubscribe(t.name, t.queueGroup, stanHandler, stan.StartAt(startPos))
	}

	if err != nil {
		return nil, err
	}
	return sub, nil
}

// ----------------------------------------------------------------------------

// prepareConnectionArgs parses args for NewSubscriber/NewPublisher from URL.
//
// TODO: maybe better re-do NewSubscriber/NewPublisher on their own to do this?
func prepareConnectionArgs(u *url.URL) (
	natsConn *natsio.Conn,
	clusterID string,
	clientID string,
	opts []stan.Option,
	err error,
) {
	q := u.Query()

	// common details/options:
	clusterID = strings.Trim(u.Path, "/")
	if clientID = q.Get("client_id"); clientID == "" {
		clientID = bps.GenClientID()
	}

	// managed NATS connection, for TLS etc:

	natsURL := &url.URL{
		Scheme: "nats",
		Host:   u.Host, // host or host:port
	}

	var natsOpts []natsio.Option
	if clientCert, clientKey := q.Get("client_cert"), q.Get("client_key"); clientCert != "" || clientKey != "" {
		if clientCert == "" {
			return nil, "", "", nil, errors.New("no client_cert provided")
		}
		if clientKey == "" {
			return nil, "", "", nil, errors.New("no client_key provided")
		}
		natsOpts = append(natsOpts, natsio.ClientCert(clientCert, clientKey))
	}

	natsConn, err = natsio.Connect(natsURL.String(), natsOpts...)
	if err != nil {
		return nil, "", "", nil, err
	}

	opts = append(opts, stan.NatsConn(natsConn))

	return
}
