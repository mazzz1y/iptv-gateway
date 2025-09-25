package demux

import (
	"io"
	"sync"
)

type streamReader struct {
	*io.PipeReader
	closeFunc func()
	once      sync.Once
}

func (sr *streamReader) Close() error {
	var err error
	sr.once.Do(func() {
		if sr.closeFunc != nil {
			sr.closeFunc()
		}
		err = sr.PipeReader.Close()
	})
	return err
}
