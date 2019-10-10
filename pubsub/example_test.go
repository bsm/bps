package pubsub_test

import (
	"context"

	"github.com/bsm/bps"
	_ "github.com/bsm/bps/pubsub"
)

func Example() {
	ctx := context.TODO()
	pub, err := bps.NewPublisher(ctx, "pubsub://my-project-id/")
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	pub.Topic("topic").Publish(ctx, &bps.Message{Data: []byte("message")})
}
