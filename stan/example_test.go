package stan_test

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/bps"
	_ "github.com/bsm/bps/stan"
)

func ExamplePublisher() {
	ctx := context.TODO()
	pub, err := bps.NewPublisher(ctx, "stan://"+natsAddr+"/"+clusterID+"?client_id=my_client")
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	pub.Topic("topic").Publish(ctx, &bps.PubMessage{Data: []byte("message")})
}

func ExampleSubscriber() {
	subscriber, err := bps.NewSubscriber(context.TODO(), "stan://"+natsAddr+"/"+clusterID+"?client_id=my_client&start_at=first")
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
