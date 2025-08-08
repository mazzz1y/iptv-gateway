package ioutils

import (
	"io"
	"sync"
)

const (
	clientBufferSize = 64
)

type AsyncWriter struct {
	client     io.Writer
	dataChan   chan []byte
	doneChan   chan struct{}
	statsLock  sync.Mutex
	droppedMsg int
}

func NewAsyncWriter(client io.Writer) *AsyncWriter {
	cw := &AsyncWriter{
		client:   client,
		dataChan: make(chan []byte, clientBufferSize),
		doneChan: make(chan struct{}),
	}

	go cw.writeLoop()
	return cw
}

func (cw *AsyncWriter) writeLoop() {
	defer close(cw.doneChan)

	for data := range cw.dataChan {
		_, err := cw.client.Write(data)
		if err != nil {
			break
		}
	}
}

func (cw *AsyncWriter) Write(data []byte) {
	buf := make([]byte, len(data))
	copy(buf, data)

	select {
	case cw.dataChan <- buf:
		return
	default:
	}

	select {
	case <-cw.dataChan:
		cw.statsLock.Lock()
		cw.droppedMsg++
		cw.statsLock.Unlock()
	default:
	}

	select {
	case cw.dataChan <- buf:
	default:
		cw.statsLock.Lock()
		cw.droppedMsg++
		cw.statsLock.Unlock()
	}
}

func (cw *AsyncWriter) Close() {
	close(cw.dataChan)
	<-cw.doneChan
}
