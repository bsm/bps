package nats_test

import (
	"context"
	"fmt"

	"github.com/bsm/bps"
	_ "github.com/bsm/bps/nats"
)

func ExamplePublisher() {
	ctx := context.TODO()
	pub, err := bps.NewPublisher(ctx, "nats://localhost:4222/?client_id=my_client")
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	pub.Topic("topic").Publish(ctx, &bps.PubMessage{Data: []byte("message")})
}

func ExampleSubscriber() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	subscriber, err := bps.NewSubscriber(ctx, "nats://localhost:4222/?client_id=my_client&start_at=first")
	if err != nil {
		panic(err.Error())
	}
	defer subscriber.Close()

	var subscription bps.Subscription
	subscription, err = subscriber.Topic("topic").Subscribe(
		bps.HandlerFunc(func(msg bps.SubMessage) {
			_, _ = fmt.Printf("%s\n", string(msg.Data()))

			// recipe to abort everything on irrecoverable handler errors:
			_ = subscription.Close() // abort subscription, so current message won't be Ack-ed
			cancel()                 // and cancel parent/program context to finally exit
		}),
		bps.StartAt(bps.PositionOldest),
	)
	if err != nil {
		panic(err.Error())
	}
	defer subscription.Close()

	<-ctx.Done() // block till subscription errors

	// if interested in subscription error:
	// this call will wait till subscription is stopped,
	// and then it'll return last handler error (mostly failed Ack, if any):
	if err := subscription.Close(); err != nil {
		panic(err.Error())
	}
}
