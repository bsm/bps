package google_test

import (
	"context"

	"github.com/bsm/bps"
	_ "github.com/bsm/bps/google"
)

func Example() {
	ctx := context.TODO()
	pub, err := bps.NewPublisher(ctx, "google://my-project-id/")
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	pub.Topic("topic").Publish(ctx, &bps.Message{Data: []byte("message")})
}
