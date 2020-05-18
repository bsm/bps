package kafka_test

import (
	"context"

	"github.com/bsm/bps"
	_ "github.com/bsm/bps/kafka"
)

func Example() {
	ctx := context.TODO()
	pub, err := bps.NewPublisher(ctx, "kafka://10.0.0.1:9092,10.0.0.2:9092,10.0.0.3:9092/?client.id=my-client&kafka.version=2.3.0&channel.buffer.size=1024")
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	pub.Topic("topic").Publish(ctx, &bps.PubMessage{Data: []byte("message")})
}
