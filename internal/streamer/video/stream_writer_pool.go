package video

import (
	"context"
	"io"
	"iptv-gateway/internal/logging"
	"sync"
)

type WriterPool struct {
	writers map[string]*StreamWriter
	mutex   sync.Mutex
	doneCh  chan struct{}
}

func NewWriterPool() *WriterPool {
	pool := &WriterPool{
		writers: make(map[string]*StreamWriter),
		doneCh:  make(chan struct{}),
	}

	return pool
}

func (p *WriterPool) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	select {
	case <-p.doneCh:
	default:
		close(p.doneCh)
	}

	for key, writer := range p.writers {
		writer.Close()
		delete(p.writers, key)
	}
}

func (p *WriterPool) AddClient(streamKey string, client io.Writer) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	writer, exists := p.writers[streamKey]
	if !exists {
		writer = NewStreamWriter()
		p.writers[streamKey] = writer
	}

	writer.AddClient(client)
	return !exists
}

func (p *WriterPool) RemoveClient(streamKey string, client io.Writer) {
	p.mutex.Lock()
	writer, exists := p.writers[streamKey]
	p.mutex.Unlock()

	if exists {
		writer.RemoveClient(client)
	}
}

func (p *WriterPool) GetWriter(streamKey string) *StreamWriter {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.writers[streamKey]
}

func (p *WriterPool) Cleanup(ctx context.Context) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var keysToRemove []string
	for key, writer := range p.writers {
		if writer.IsEmpty() {
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		delete(p.writers, key)
		logging.Debug(ctx, "removed empty stream")
	}
}
