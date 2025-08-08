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
	mutex, _ := m.streamLocks.LoadOrStore(streamKey, &sync.Mutex{})
	mtx := mutex.(*sync.Mutex)

	mtx.Lock()

	return func() {
		mtx.Unlock()
	}
}

func (m *StreamManager) GetStream(req StreamRequest) (io.ReadCloser, error) {
	unlock := m.LockStream(req.StreamKey)
	defer unlock()

	pr, pw := io.Pipe()

	ctx := context.WithValue(req.Context, constant.ContextStreamID, req.StreamKey)

	go func() {
		<-ctx.Done()
		pr.Close()
		pw.Close()
		m.pool.RemoveClient(req.StreamKey, pw)
		logging.Debug(ctx, "context done, reader closed")
	}()

	isNewStream := m.pool.AddClient(req.StreamKey, pw)
	if isNewStream {
		go m.startStream(ctx, req, pw)
		logging.Info(ctx, "started new stream")
	} else {
		logging.Info(ctx, "joined existing stream")
	}

	return pr, nil
}

func (m *StreamManager) startStream(ctx context.Context, req StreamRequest, w io.Writer) {
	key := req.StreamKey

	unlock := m.LockStream(key)
	writer := m.pool.GetWriter(key)
	unlock()

	if writer == nil {
		logging.Error(ctx, nil, "failed to get writer")
		if closer, ok := w.(io.Closer); ok {
			closer.Close()
		}
		return
	}

	streamID := ctx.Value(constant.ContextStreamID).(string)
	streamCtx, cancel := context.WithCancel(
		context.WithValue(context.Background(), constant.ContextStreamID, streamID))
	defer cancel()

	go func() {
		emptyCh := writer.IsEmptyChannel()
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
