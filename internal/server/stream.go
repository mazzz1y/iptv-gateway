package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"time"

	"iptv-gateway/internal/app"
	"iptv-gateway/internal/ctxutil"
	"iptv-gateway/internal/demux"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/metrics"
	"iptv-gateway/internal/urlgen"
)

const (
	maxRetryAttempts = 2
	retryTimeout     = 3 * time.Second
)

type streamResult struct {
	success         bool
	isLimitError    bool
	isUpstreamError bool
}

type allStreamsResult struct {
	success          bool
	hasLimitError    bool
	hasUpstreamError bool
	defaultProvider  *app.Playlist
}

func (s *Server) handleStreamProxy(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logging.Debug(ctx, "proxying stream")

	client := ctxutil.Client(ctx).(*app.Client)
	data := ctxutil.StreamData(ctx).(*urlgen.Data)

	ctx = ctxutil.WithChannelName(ctx, data.StreamData.ChannelName)

	if !s.acquireSemaphores(ctx) {
		logging.Error(ctx, errors.New("failed to acquire semaphores"), "")
		if len(data.StreamData.Streams) > 0 {
			firstProvider := client.GetProvider(
				data.StreamData.Streams[0].ProviderInfo.ProviderType,
				data.StreamData.Streams[0].ProviderInfo.ProviderName,
			)
			if playlist, ok := firstProvider.(*app.Playlist); ok && playlist != nil {
				playlist.LimitStreamer().Stream(ctx, w)
				return
			}
		}
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.releaseSemaphores(ctx)

	var lastResult allStreamsResult

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if attempt > 0 && lastResult.hasLimitError {
			logging.Debug(ctx, "sleeping before retry attempt")
			time.Sleep(retryTimeout)
		}

		result := s.tryAllStreams(ctx, w, r, client, data)
		if result.success {
			return
		}

		lastResult = result
	}

	if lastResult.hasLimitError && lastResult.defaultProvider != nil {
		lastResult.defaultProvider.LimitStreamer().Stream(ctx, w)
		return
	}

	if lastResult.hasUpstreamError && lastResult.defaultProvider != nil {
		lastResult.defaultProvider.UpstreamErrorStreamer().Stream(ctx, w)
		return
	}

	http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
}

func (s *Server) tryAllStreams(
	ctx context.Context, w http.ResponseWriter, r *http.Request, client *app.Client, data *urlgen.Data) allStreamsResult {
	var hasLimitError bool
	var hasUpstreamError bool
	var firstProvider *app.Playlist

	for i, stream := range data.StreamData.Streams {
		logging.Debug(ctx, "trying stream source", "index", i)

		provider := client.GetProvider(stream.ProviderInfo.ProviderType, stream.ProviderInfo.ProviderName)
		if provider == nil {
			logging.Error(ctx, errors.New("provider not found"), "stream_index", i,
				"provider_type", stream.ProviderInfo.ProviderType,
				"provider_name", stream.ProviderInfo.ProviderName)
			continue
		}

		playlist, ok := provider.(*app.Playlist)
		if !ok {
			logging.Error(ctx, errors.New("provider is not a playlist"), "stream_index", i)
			continue
		}

		if firstProvider == nil {
			firstProvider = playlist
		}

		result := s.tryStream(ctx, w, r, playlist, stream, i)

		if result.success {
			return allStreamsResult{true, false, false, nil}
		}

		if result.isLimitError {
			hasLimitError = true
		} else if result.isUpstreamError {
			hasUpstreamError = true
		}
	}

	return allStreamsResult{
		false, hasLimitError, hasUpstreamError, firstProvider}
}

func (s *Server) tryStream(
	ctx context.Context,
	w http.ResponseWriter, r *http.Request,
	playlist *app.Playlist, stream urlgen.Stream, streamIndex int) streamResult {

	ctx = ctxutil.WithChannelHidden(ctx, stream.Hidden)

	streamURL := buildStreamURL(stream.URL, r.URL.RawQuery)
	streamKey := buildStreamKey(stream.URL, r.URL.RawQuery)

	streamSource := playlist.LinkStreamer(streamURL)
	if streamSource == nil {
		logging.Error(ctx,
			errors.New("failed to create stream source"), "stream_index", streamIndex)

		return streamResult{false, false, false}
	}

	demuxReq := demux.Request{
		StreamKey: streamKey,
		Streamer:  streamSource,
		Semaphore: playlist.Semaphore(),
	}

	reader, err := s.demux.GetReader(ctx, demuxReq)
	if errors.Is(err, demux.ErrSubscriptionSemaphore) {
		s.handleSubscriptionError(ctx, playlist, stream, streamIndex)
		return streamResult{false, true, false}
	}
	if err != nil {
		logging.Error(ctx, err, "failed to get stream", "stream_index", streamIndex)
		return streamResult{false, false, false}
	}
	defer reader.Close()

	logging.Debug(ctx, "started stream", "stream_index", streamIndex)
	result := s.streamToResponse(ctx, w, reader, playlist, stream)
	return result
}

func (s *Server) handleSubscriptionError(
	ctx context.Context,
	playlist *app.Playlist, stream urlgen.Stream, streamIndex int) {

	logging.Error(
		ctx, demux.ErrSubscriptionSemaphore,
		"failed to get stream - subscription semaphore", "stream_index", streamIndex)

	if !stream.Hidden {
		data := ctxutil.StreamData(ctx).(*urlgen.Data)
		clientName := ctxutil.ClientName(ctx)
		playlistName := playlist.Name()
		metrics.StreamsFailuresTotal.WithLabelValues(
			clientName, playlistName, data.StreamData.ChannelName, metrics.FailureReasonPlaylistLimit,
		).Inc()
	}
}

func (s *Server) streamToResponse(
	ctx context.Context,
	w http.ResponseWriter, reader io.ReadCloser, playlist *app.Playlist, stream urlgen.Stream) streamResult {

	data := ctxutil.StreamData(ctx).(*urlgen.Data)
	clientName := ctxutil.ClientName(ctx)
	playlistName := playlist.Name()

	if !stream.Hidden {
		metrics.ClientStreamsActive.WithLabelValues(clientName, playlistName, data.StreamData.ChannelName).Inc()
		defer metrics.ClientStreamsActive.WithLabelValues(clientName, playlistName, data.StreamData.ChannelName).Dec()
	}

	w.Header().Set("Content-Type", streamContentType)
	written, err := io.Copy(w, reader)

	if err == nil && written == 0 {
		if !stream.Hidden {
			logging.Error(ctx, errors.New("no data written to response"), "")
			metrics.StreamsFailuresTotal.WithLabelValues(
				clientName, playlistName, data.StreamData.ChannelName, metrics.FailureReasonUpstreamError,
			).Inc()
		}
		return streamResult{false, false, true}
	}

	if err != nil && !isClientDisconnect(err) {
		logging.Error(ctx, err, "error copying stream to response")
		fmt.Printf("err type = %T\n", err)
	}

	return streamResult{true, false, false}
}

func isClientDisconnect(err error) bool {
	return errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNRESET)
}
