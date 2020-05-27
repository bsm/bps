package kafka_test

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/bps"
	_ "github.com/bsm/bps/kafka"
)

func ExamplePublisher() {
	ctx := context.TODO()
	pub, err := bps.NewPublisher(ctx, "kafka://10.0.0.1:9092,10.0.0.2:9092,10.0.0.3:9092/?client.id=my-client&kafka.version=2.3.0&channel.buffer.size=1024")
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	pub.Topic("topic").Publish(ctx, &bps.PubMessage{Data: []byte("message")})
}

func ExampleSubscriber() {
	subscriber, err := bps.NewSubscriber(context.TODO(), "kafka://10.0.0.1:9092,10.0.0.2:9092,10.0.0.3:9092/?client.id=my-client&kafka.version=2.3.0&channel.buffer.size=1024")
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
