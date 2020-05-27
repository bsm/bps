package bps_test

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/bps"
)

func ExamplePublisher() {
	ctx := context.Background()
	pub := bps.NewInMemPublisher()
	defer pub.Close()

	topicA := pub.Topic("topic-a")
	topicB := pub.Topic("topic-b")

	topicA.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-1"),
	})
	topicB.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-2"),
	})
	topicA.Publish(ctx, &bps.PubMessage{
		Data: []byte("message-2"),
	})

	fmt.Println(len(topicA.(*bps.InMemPubTopic).Messages()))
	fmt.Println(len(topicB.(*bps.InMemPubTopic).Messages()))

	// Output:
	// 2
	// 1
}

func ExampleSubscriber() {
	subscriber := bps.NewInMemSubscriber(
		map[string][]bps.SubMessage{
			"foo": []bps.SubMessage{
				bps.RawSubMessage("foo1"),
				bps.RawSubMessage("foo2"),
			},
		},
	)
	defer subscriber.Close()

	subscription, err := subscriber.Topic("foo").Subscribe(
		bps.HandlerFunc(func(msg bps.SubMessage) {
			fmt.Printf("%s\n", msg.Data())
		}),
	)
	if err != nil {
		panic(err.Error())
	}
	defer subscription.Close()

	time.Sleep(time.Second) // wait to receive some messages

	// Output:
	// foo1
	// foo2
}
