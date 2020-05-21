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
// bps.NewSubscriber supports:
//
//   start_at
//     new_only - consume messages, published after subscription
//     first - consume from first/oldest available message
//
package nats

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bsm/bps"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

func init() {
	bps.RegisterPublisher("nats", func(ctx context.Context, u *url.URL) (bps.Publisher, error) {
		return NewConn(parseConnectionParams(u))
	})
	bps.RegisterSubscriber("nats", func(ctx context.Context, u *url.URL) (bps.Subscriber, error) {
		return NewConn(parseConnectionParams(u))
	})
}

// Conn is a wrapper for stan.Conn, that implements both bps.Publisher and bps.Subscriber interfaces.
type Conn struct {
	stan    stan.Conn
	subOpts []stan.SubscriptionOption
}

// NewConn constructs a new nats.io-backed pub/sub connection.
func NewConn(stanClusterID, clientID string, opts []stan.Option, subOpts []stan.SubscriptionOption) (*Conn, error) {
	c, err := stan.Connect(stanClusterID, clientID, opts...)
	if err != nil {
		return nil, err
	}

	return &Conn{
		stan: c,
		subOpts: append(
			append(make([]stan.SubscriptionOption, 0, len(subOpts)+1), subOpts...), // copy provided opts
			stan.SetManualAckMode(), // force manual ack mode, it's handled by this impl
		),
	}, nil
}

// Topic returns producer topic.
func (c *Conn) Topic(name string) bps.Topic {
	return &topic{
		stan: c.stan,
		name: name,
	}
}

// Subscribe subscribes to topic messages.
func (c *Conn) Subscribe(ctx context.Context, topic string, handler bps.Handler) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// atomic flag to stop after first handler error returned;
	// client received (buffers?) messages in background,
	// may call handler after it returns error:
	var done int32

	var (
		sub stan.Subscription
		err error
	)

	stop := func(cause error) {
		// stop after first handler error returned:
		atomic.StoreInt32(&done, 1)

		// cancel context to return from Subscribe:
		cancel()

		// store error error to return:
		if err == nil {
			err = cause
		}
	}

	sub, err = c.stan.Subscribe(topic, func(msg *stan.Msg) {
		// stop after first handler error returned:
		if atomic.LoadInt32(&done) != 0 {
			return
		}

		if err := handler.Handle(bps.RawSubMessage(msg.Data)); errors.Is(err, bps.Done) {
			// stop normally, still acknowledge message:
			stop(nil)
		} else if err != nil {
			// stop with error, do not acknowledge message:
			stop(err)
			return
		}

		if err := msg.Ack(); err != nil {
			stop(err)
			return
		}
	}, c.subOpts...)
	if err != nil {
		return err
	}
	defer sub.Close()

	<-ctx.Done()
	return err
}

// Close terminates connection.
func (c *Conn) Close() error {
	return c.stan.Close()
}

// ----------------------------------------------------------------------------

type topic struct {
	stan stan.Conn
	name string
}

func (t *topic) Publish(ctx context.Context, msg *bps.PubMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return t.stan.Publish(t.name, msg.Data)
}

func (t *topic) PublishBatch(ctx context.Context, messages []*bps.PubMessage) error {
	for i, msg := range messages {
		if err := t.Publish(ctx, msg); err != nil {
			return fmt.Errorf("publish %d of %d: %w", i, len(messages), err)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

func parseConnectionParams(u *url.URL) (
	clusterID string,
	clientID string,
	opts []stan.Option,
	subOpts []stan.SubscriptionOption,
) {
	q := u.Query()

	// common details/options:

	clusterID = strings.Trim(u.Path, "/")

	if clientID = q.Get("client_id"); clientID == "" {
		clientID = genClientID()
	}

	opts = append(opts, stan.NatsURL((&url.URL{
		Scheme: "nats",
		Host:   u.Host, // host or host:port
	}).String()))

	// subscriber options:

	switch q.Get("start_at") {
	case "new_only":
		subOpts = append(subOpts, stan.StartAt(pb.StartPosition_NewOnly))
	case "first":
		subOpts = append(subOpts, stan.StartAt(pb.StartPosition_First))
	}

	return
}

func genClientID() string {
	return fmt.Sprintf("bps-%d", time.Now().UnixNano())
}
