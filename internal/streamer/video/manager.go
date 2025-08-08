package video

import (
	"context"
	"errors"
	"io"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/logging"
	"sync"
)

type StreamSource interface {
	Stream(ctx context.Context, w io.Writer) (int64, error)
}

type StreamRequest struct {
	Context    context.Context
	StreamKey  string
	StreamData StreamSource
}

type StreamManager struct {
	pool          *WriterPool
	rootCtx       context.Context
	rootCtxCancel context.CancelFunc
	streamLocks   sync.Map
}

func NewStreamManager() *StreamManager {
	rootCtx, cancel := context.WithCancel(context.Background())
	return &StreamManager{
		pool:          NewWriterPool(),
		rootCtx:       rootCtx,
		rootCtxCancel: cancel,
		streamLocks:   sync.Map{},
	}
}

func (m *StreamManager) Stop() {
	if m.rootCtxCancel != nil {
		m.rootCtxCancel()
	}
	m.pool.Stop()
}

func (m *StreamManager) HasActiveStream(streamKey string) bool {
	writer := m.pool.GetWriter(streamKey)
	return writer != nil && !writer.IsEmpty()
}

func (m *StreamManager) LockStream(streamKey string) func() {
	actualMutex, _ := m.streamLocks.LoadOrStore(streamKey, &sync.Mutex{})
	mutex := actualMutex.(*sync.Mutex)
	mutex.Lock()

	return func() {
		mutex.Unlock()
	}
}

func (m *StreamManager) CreateStream(req StreamRequest) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	ctx := context.WithValue(req.Context, constant.ContextStreamID, req.StreamKey)

	go func() {
		<-ctx.Done()
		pr.Close()
		logging.Debug(ctx, "context done, reader closed")
	}()

	isNewStream := m.pool.AddClient(req.StreamKey, pw)

	if isNewStream {
		go m.startStreamWithWriter(ctx, req, pw)
		logging.Info(ctx, "started new stream")
	} else {
		logging.Info(ctx, "joined existing stream")
	}

	return pr, nil
}

func (m *StreamManager) GetStream(req StreamRequest) (io.ReadCloser, error) {
	if m.HasActiveStream(req.StreamKey) {
		return m.CreateStream(req)
	}

	unlock := m.LockStream(req.StreamKey)
	defer unlock()

	if m.HasActiveStream(req.StreamKey) {
		return m.CreateStream(req)
	}

	return m.CreateStream(req)
}

// ServeStream creates or gets a stream and returns a reader for it
func (m *StreamManager) ServeStream(req StreamRequest) (io.ReadCloser, error) {
	return m.GetStream(req)
}

func (m *StreamManager) startStreamWithWriter(ctx context.Context, req StreamRequest, w io.Writer) {
	key := req.StreamKey
	writer := m.pool.GetWriter(key)

	if writer == nil {
		logging.Error(ctx, nil, "failed to get writer")
		if closer, ok := w.(io.Closer); ok {
			closer.Close()
		}
		return
	}

	detachedCtx := m.rootCtx
	if streamID, ok := ctx.Value(constant.ContextStreamID).(string); ok {
		detachedCtx = context.WithValue(detachedCtx, constant.ContextStreamID, streamID)
	}

	streamCtx, cancel := context.WithCancel(detachedCtx)
	defer cancel()

	go func() {
		emptyCh := writer.NotifyWhenEmpty()
		defer writer.CancelNotify(emptyCh)

		select {
		case <-emptyCh:
			logging.Debug(ctx, "no clients left, stopping stream")
			cancel()
			return
		case <-streamCtx.Done():
			logging.Debug(ctx, "context canceled, stopping stream")
			return
		}
	}()

	logging.Info(ctx, "starting stream")
	bytesWritten, err := req.StreamData.Stream(streamCtx, writer)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			logging.Info(streamCtx, "stream canceled")
		} else {
			logging.Error(streamCtx, err, "stream failed")
		}
	} else if bytesWritten == 0 {
		logging.Error(streamCtx, nil, "stream produced no output")
	} else {
		logging.Info(streamCtx, "stream ended")
	}

	m.pool.Cleanup(streamCtx)

	if closer, ok := w.(io.Closer); ok {
		closer.Close()
	}
}
