package ioutils

import (
	"errors"
	"io"
	"time"
)

var ErrReaderTimeout = errors.New("timeout")

type TimeoutReader struct {
	r       io.Reader
	timeout time.Duration
}

func NewTimeoutReader(r io.Reader, timeout time.Duration) *TimeoutReader {
	return &TimeoutReader{r: r, timeout: timeout}
}

func (r *TimeoutReader) Read(p []byte) (int, error) {
	c := make(chan struct{})
	var n int
	var err error

	go func() {
		n, err = r.r.Read(p)
		close(c)
	}()

	select {
	case <-c:
		return n, err
	case <-time.After(r.timeout):
		return 0, ErrReaderTimeout
	}
}
