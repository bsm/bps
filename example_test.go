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

	topicA.Publish(ctx, &bps.Message{
		Data: []byte("message-1"),
	})
	topicB.Publish(ctx, &bps.Message{
		Data: []byte("message-2"),
	})
	topicA.Publish(ctx, &bps.Message{
		Data: []byte("message-2"),
	})

	fmt.Println(len(topicA.(*bps.InMemTopic).Messages()))
	fmt.Println(len(topicB.(*bps.InMemTopic).Messages()))

	// Output:
	// 2
	// 1
}
