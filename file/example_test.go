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

func Example() {
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
