package ioutil

import (
	"io"
	"sync"
)

type ReaderWithCloser struct {
	reader io.Reader
	closer func() error
	mu     sync.Mutex
	closed bool
}

func NewReaderWithCloser(reader io.Reader, closer func() error) *ReaderWithCloser {
	return &ReaderWithCloser{reader: reader, closer: closer}
}

func (w *ReaderWithCloser) Read(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed || w.reader == nil {
		return 0, io.EOF
	}
	return w.reader.Read(p)
}

func (w *ReaderWithCloser) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	if w.closer != nil {
		return w.closer()
	}
	return nil
}
