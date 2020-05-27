package file_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bsm/bps"
	_ "github.com/bsm/bps/file"
)

func ExamplePublisher() {
	dir, err := ioutil.TempDir("", "bps-example")
	if err != nil {
		panic(err.Error())
	}
	defer os.RemoveAll(dir)

	ctx := context.TODO()

	// create a publisher
	pub, err := bps.NewPublisher(ctx, "file://"+dir)
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	// add a message to topic
	pub.Topic("topic").Publish(ctx, &bps.PubMessage{Data: []byte("message")})

	// check files in dir
	entries, _ := filepath.Glob(dir + "/*")
	fmt.Println(len(entries))
	fmt.Println(strings.ReplaceAll(entries[0], dir, ""))

	// Output:
	// 1
	// /topic
}

func ExampleSubscriber() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir, err := ioutil.TempDir("", "bps-example")
	if err != nil {
		panic(err.Error())
	}
	defer os.RemoveAll(dir)

	// produce some messages (seed topic):
	SeedTopic(ctx, dir, "topic")

	subscriber, err := bps.NewSubscriber(ctx, "file://"+dir)
	if err != nil {
		panic(err.Error())
	}
	defer subscriber.Close()

	subscription, err := subscriber.Topic("topic").Subscribe(
		bps.HandlerFunc(func(msg bps.SubMessage) {
			fmt.Printf("%s\n", msg.Data())
			cancel() // handle only one message and signal program to exit
		}),
	)
	if err != nil {
		panic(err.Error())
	}
	defer subscription.Close()

	<-ctx.Done() // wait till message is received

	// if interested in subscription error:
	// this call will wait till subscription is stopped,
	// and then it'll return last handler error (mostly failed Ack, if any):
	if err := subscription.Close(); err != nil {
		panic(err.Error())
	}

	// Output:
	// message
}

func SeedTopic(ctx context.Context, dir, topic string) {
	// create a publisher
	pub, err := bps.NewPublisher(ctx, "file://"+dir)
	if err != nil {
		panic(err.Error())
	}
	defer pub.Close()

	// add a message to topic
	pub.Topic("topic").Publish(ctx, &bps.PubMessage{Data: []byte("message")})

	// flush/finalize publisher - files should not be written when they are consumed:
	if err := pub.Close(); err != nil {
		panic(err.Error())
	}
}
