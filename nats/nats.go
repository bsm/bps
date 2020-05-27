// Package nats abstracts publish-subscribe https://nats.io backend.
//
// WARNING/TODO: messages can be considered not acknowledged,
//               because DurableName for subscriptions is not used yet:
//               https://godoc.org/github.com/nats-io/stan.go#DurableName
//
// Both bps.NewPublisher and bps.NewSubscriber support:
//
//   client_id
//     nats-streaming client ID, [0-9A-Za-z_-] only.
//
package nats

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/bsm/bps"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

func init() {
	bps.RegisterPublisher("nats", func(ctx context.Context, u *url.URL) (bps.Publisher, error) {
		return NewPublisher(parseConnectionParams(u))
	})
	bps.RegisterSubscriber("nats", func(ctx context.Context, u *url.URL) (bps.Subscriber, error) {
		return NewSubscriber(parseConnectionParams(u))
	})
}

type publisher struct {
	conn stan.Conn
}

// NewPublisher constructs a new nats.io-backed publisher.
func NewPublisher(stanClusterID, clientID string, opts []stan.Option) (bps.Publisher, error) {
	c, err := stan.Connect(stanClusterID, clientID, opts...)
	if err != nil {
		return nil, err
	}

	return &publisher{conn: c}, nil
}

func (p *publisher) Topic(name string) bps.PubTopic {
	return &pubTopic{
		conn: p.conn,
		name: name,
	}
}

func (p *publisher) Close() error {
	return p.conn.Close()
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
	conn stan.Conn
}

// NewSubscriber constructs a new nats.io-backed publisher.
// By default, it starts handling from the newest available message (published after subscribing).
func NewSubscriber(stanClusterID, clientID string, opts []stan.Option) (bps.Subscriber, error) {
	c, err := stan.Connect(stanClusterID, clientID, opts...)
	if err != nil {
		return nil, err
	}

	return &subscriber{conn: c}, nil
}

func (s *subscriber) Topic(name string) bps.SubTopic {
	return &subTopic{
		stan: s.conn,
		name: name,
	}
}

func (s *subscriber) Close() error {
	return s.conn.Close()
}

// ----------------------------------------------------------------------------

type subTopic struct {
	stan stan.Conn
	name string
}

func (t *subTopic) Subscribe(ctx context.Context, handler bps.Handler, options ...bps.SubOption) error {
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
		return fmt.Errorf("start position %s is not supported by this implementation", opts.StartAt)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		sub stan.Subscription
		err error
	)

	stop := func(cause error) {
		// cancel context to return from Subscribe:
		cancel()

		// store error error to return:
		if err == nil {
			err = cause
		}
	}

	sub, err = t.stan.Subscribe(
		t.name,
		func(msg *stan.Msg) {
			// stop after first handler error returned:
			// stan.Conn may still call handler to process buffered (?) messages:
			select {
			case <-ctx.Done():
				return
			default:
			}

			handler.Handle(bps.RawSubMessage(msg.Data))

			if err := msg.Ack(); err != nil {
				stop(err)
				return
			}
		},
		stan.SetManualAckMode(), // force manual ack mode, it's handled by this impl
		stan.StartAt(startPos),
	)
	if err != nil {
		return err
	}
	defer sub.Close()

	<-ctx.Done()
	return err
}

// ----------------------------------------------------------------------------

func parseConnectionParams(u *url.URL) (
	clusterID string,
	clientID string,
	opts []stan.Option,
) {
	q := u.Query()

	// common details/options:

	clusterID = strings.Trim(u.Path, "/")

	if clientID = q.Get("client_id"); clientID == "" {
		clientID = bps.GenClientID()
	}

	opts = append(opts, stan.NatsURL((&url.URL{
		Scheme: "nats",
		Host:   u.Host, // host or host:port
	}).String()))

	return
}
