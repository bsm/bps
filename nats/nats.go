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
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bsm/bps"
	"github.com/nats-io/stan.go"
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
	stan stan.Conn
}

// NewConn constructs a new nats.io-backed pub/sub connection.
func NewConn(stanClusterID, clientID string, opts []stan.Option) (*Conn, error) {
	c, err := stan.Connect(stanClusterID, clientID, opts...)
	if err != nil {
		return nil, err
	}

	return &Conn{
		stan: c,
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

	sub, err = c.stan.Subscribe(
		topic,
		func(msg *stan.Msg) {
			// stop after first handler error returned:
			// stan.Conn may still call handler to process buffered (?) messages:
			select {
			case <-ctx.Done():
				return
			default:
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
		},
		stan.SetManualAckMode(),    // force manual ack mode, it's handled by this impl
		stan.DeliverAllAvailable(), // TODO: make it configurable (initial offset = oldest/newest; DeliverAllAvailable -> oldest)
	)
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

	return
}

func genClientID() string {
	return fmt.Sprintf("bps-%d", time.Now().UnixNano())
}
