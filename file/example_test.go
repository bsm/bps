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
	dir, err := ioutil.TempDir("", "bps-example")
	if err != nil {
		panic(err.Error())
	}
	defer os.RemoveAll(dir)

	ctx := context.TODO()

	// produce some messages (seed topic):
	SeedTopic(ctx, dir, "topic")

	sub, err := bps.NewSubscriber(ctx, "file://"+dir)
	if err != nil {
		panic(err.Error())
	}
	defer sub.Close()

	handler := bps.HandlerFunc(func(msg bps.SubMessage) error {
		fmt.Printf("%s\n", msg.Data())
		return nil // or bps.Done to stop/unsubscribe
	})

	// blocks till all the messages are consumed or error occurs:
	err = sub.Topic("topic").Subscribe(ctx, handler)
	if err != nil {
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
