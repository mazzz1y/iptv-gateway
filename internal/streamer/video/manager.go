package video

import (
	"context"
	"errors"
	"golang.org/x/sync/semaphore"
	"io"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/utils"
	"sync"
)

var (
	ErrSubscriptionSemaphore = errors.New("failed to acquire subscription semaphore")
)

type StreamSource interface {
	Stream(ctx context.Context, w io.Writer) (int64, error)
}

type StreamRequest struct {
	Context    context.Context
	StreamKey  string
	StreamData StreamSource
	Semaphore  *semaphore.Weighted
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

func (m *StreamManager) LockStream(streamKey string) func() {
	mutex, _ := m.streamLocks.LoadOrStore(streamKey, &sync.Mutex{})
	mtx := mutex.(*sync.Mutex)

	mtx.Lock()

	return func() {
		mtx.Unlock()
	}
}

func (m *StreamManager) GetReader(req StreamRequest) (io.ReadCloser, error) {
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
		if utils.AcquireSemaphore(ctx, req.Semaphore, "subscription") {
			logging.Debug(ctx, "acquired subscription semaphore")
		} else {
			return nil, ErrSubscriptionSemaphore
		}
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

		if req.Semaphore != nil {
			defer req.Semaphore.Release(1)
			defer logging.Debug(ctx, "releasing subscription semaphore")
		}

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

	if closer, ok := w.(io.Closer); ok {
		closer.Close()
	}
}
