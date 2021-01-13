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
	pub, err := bps.NewPublisher(ctx, "nats://"+natsAddr)
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	pub.Topic("topic").Publish(ctx, &bps.PubMessage{Data: []byte("message")})
}

func ExampleSubscriber() {
	subscriber, err := bps.NewSubscriber(context.TODO(), "nats://"+natsAddr)
	if err != nil {
		panic(err.Error())
	}
	defer subscriber.Close()

	subscription, err := subscriber.Topic("topic").Subscribe(
		bps.HandlerFunc(func(msg bps.SubMessage) {
			_, _ = fmt.Printf("%s\n", string(msg.Data()))
		}),
		bps.StartAt(bps.PositionOldest),
	)
	if err != nil {
		panic(err.Error())
	}
	defer subscription.Close()

	time.Sleep(time.Second) // wait to receive some messages
}
