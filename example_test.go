package bps_test

import (
	"context"
	"fmt"

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
	// ubsubscribing is done by canceling context, optional:
	ctx, unsubscribe := context.WithCancel(context.Background())
	defer unsubscribe()

	sub := bps.NewInMemSubscriber(
		map[string][]bps.SubMessage{
			"foo": []bps.SubMessage{
				bps.RawSubMessage("foo1"),
				bps.RawSubMessage("foo2"),
			},
		},
	)
	defer sub.Close()

	handler := bps.HandlerFunc(func(msg bps.SubMessage) error {
		fmt.Printf("%s\n", msg.Data())
		return nil // or bps.Done to stop/unsubscribe
	})

	// blocks till all the messages are consumed or error occurs:
	err := sub.Topic("foo").Subscribe(ctx, handler)
	if err != nil {
		panic(err.Error())
	}

	// Output:
	// foo1
	// foo2
}
