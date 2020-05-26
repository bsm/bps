// Package file implements a file-system producer.
package file

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/bsm/bps"
)

func init() {
	bps.RegisterPublisher("file", func(_ context.Context, u *url.URL) (bps.Publisher, error) {
		return NewPublisher(path.Join(u.Host, u.Path))
	})
	bps.RegisterSubscriber("file", func(_ context.Context, u *url.URL) (bps.Subscriber, error) {
		return NewSubscriber(path.Join(u.Host, u.Path)), nil
	})
}

// --------------------------------------------------------------------

type filePub struct {
	root string

	topics map[string]*fileTopic
	mu     sync.RWMutex
}

// NewPublisher inits a publisher within a root directory.
func NewPublisher(root string) (bps.Publisher, error) {
	if err := os.MkdirAll(root, 0777); err != nil {
		return nil, err
	}
	return &filePub{root: root, topics: make(map[string]*fileTopic)}, nil
}

func (p *filePub) Topic(name string) bps.PubTopic {
	p.mu.RLock()
	topic, ok := p.topics[name]
	p.mu.RUnlock()

	if ok {
		return topic
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if topic, ok = p.topics[name]; !ok {
		topic = &fileTopic{name: filepath.Join(p.root, name)}
		p.topics[name] = topic
	}
	return topic
}

func (p *filePub) Close() (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, topic := range p.topics {
		if e := topic.Close(); e != nil {
			err = e
		}
		delete(p.topics, name)
	}
	return
}

// --------------------------------------------------------------------

type fileTopic struct {
	name string
	file *os.File
	enc  *json.Encoder
	mu   sync.Mutex
}

func (t *fileTopic) Publish(ctx context.Context, msg *bps.PubMessage) error {
	return t.PublishBatch(ctx, []*bps.PubMessage{msg})
}

func (t *fileTopic) PublishBatch(_ context.Context, batch []*bps.PubMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.file == nil {
		file, err := os.OpenFile(t.name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		t.file = file
		t.enc = json.NewEncoder(file)
	}

	for _, msg := range batch {
		if err := t.enc.Encode(msg); err != nil {
			return err
		}
	}
	return t.file.Sync()
}

func (t *fileTopic) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.file != nil {
		return t.file.Close()
	}
	return nil
}

// --------------------------------------------------------------------

type fileSub struct {
	root string
}

// NewSubscriber inits a subscriber within a root directory.
// It assumes, that file is not being written to anymore.
// It starts handling from the first (oldest) available message.
// It iterates entire file (all records) without tracking processed messages.
func NewSubscriber(root string) bps.Subscriber {
	return &fileSub{
		root: root,
	}
}

func (s *fileSub) Topic(name string) bps.SubTopic {
	return NewSubTopic(filepath.Join(s.root, name))
}

func (s *fileSub) Close() error {
	return nil
}

// ----------------------------------------------------------------------------

type subTopic struct {
	filename string
}

// NewSubTopic inits a subscriber topic for given file name.
// Useful for testing.
func NewSubTopic(filename string) bps.SubTopic {
	return &subTopic{
		filename: filename,
	}
}

func (t *subTopic) Subscribe(ctx context.Context, handler bps.Handler, _ ...bps.SubOption) error {
	f, err := os.Open(t.filename)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	for dec.More() {
		if err := ctx.Err(); err != nil {
			return err
		}

		var msg subMessage
		if err := dec.Decode(&msg); err != nil {
			return err
		}

		if err := handler.Handle(msg); errors.Is(err, bps.Done) {
			break
		} else if err != nil {
			return err
		}
	}
	return nil
}

type subMessage struct{ bps.PubMessage }

func (m subMessage) Data() []byte { return m.PubMessage.Data }
