// Package file implements a file-system producer.
package file

import (
	"context"
	"encoding/json"
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

func (p *filePub) Topic(name string) bps.Topic {
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

func (t *fileTopic) Publish(ctx context.Context, msg *bps.Message) error {
	return t.PublishBatch(ctx, []*bps.Message{msg})
}

func (t *fileTopic) PublishBatch(_ context.Context, batch []*bps.Message) error {
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
