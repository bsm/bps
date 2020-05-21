package nats_test

import (
	"context"
	"fmt"
	"time"

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
	ctx := context.TODO()
	sub, err := bps.NewSubscriber(ctx, "nats://localhost:4222/?client_id=my_client&start_at=first")
	if err != nil {
		panic(err.Error())
	}
	defer sub.Close()

	// can be just context.WithCancel - `cancel` can be used to stop consuming:
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// will block till context is cancelled or handler returns error:
	err = sub.Subscribe(
		ctx,
		"topic",
		bps.HandlerFunc(func(msg bps.SubMessage) error {
			_, _ = fmt.Printf("%s\n", string(msg.Data()))
			return nil // or bps.Done to unsubscribe
		}),
		bps.StartAt(bps.PositionOldest),
	)
	if err != nil {
		panic(err.Error())
	}
}
