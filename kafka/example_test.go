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
	ctx := context.TODO()
	sub, err := bps.NewSubscriber(ctx, "kafka://10.0.0.1:9092,10.0.0.2:9092,10.0.0.3:9092/?client.id=my-client&kafka.version=2.3.0&channel.buffer.size=1024")
	if err != nil {
		panic(err.Error())
	}
	defer sub.Close()

	// can be just context.WithCancel - `cancel` can be used to stop consuming:
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// will block till context is cancelled:
	err = sub.Topic("topic").Subscribe(
		ctx,
		bps.HandlerFunc(func(msg bps.SubMessage) {
			_, _ = fmt.Printf("%s\n", string(msg.Data()))
		}),
		bps.StartAt(bps.PositionOldest),
	)
	if err != nil {
		panic(err.Error())
	}
}
