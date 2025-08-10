package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/manager"
	"iptv-gateway/internal/streamer/m3u8"
	"iptv-gateway/internal/streamer/video"
	"iptv-gateway/internal/streamer/xmltv"
	"iptv-gateway/internal/url_generator"
	"net/http"
	"time"
)

const (
	streamContentType = "video/mp2t"
	linkTTL           = time.Hour * 24 * 30
)

func (s *Server) handlePlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c := ctx.Value(constant.ContextClient).(*manager.Client)

	logging.Debug(ctx, "playlist request")

	w.Header().Set("Content-Type", "application/x-mpegurl")
	w.Header().Set("Cache-Control", "no-cache")

	streamer := m3u8.NewStreamer(c.GetSubscriptions(), c.GetEpgLink(), s.cache)
	if _, err := streamer.WriteTo(ctx, w); err != nil {
		logging.Error(ctx, err, "failed to write playlist")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleEPG(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logging.Info(ctx, "epg request")

	streamer, err := s.prepareEPGStreamer(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Cache-Control", "no-cache")

	if _, err = streamer.WriteTo(ctx, w); err != nil {
		logging.Error(ctx, err, "failed to write EPG")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleEPGgz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logging.Debug(ctx, "gzipped epg request")

	streamer, err := s.prepareEPGStreamer(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Disposition", "attachment; filename=\"epg.xml.gz\"")

	if _, err = streamer.WriteToGzip(ctx, w); err != nil {
		logging.Error(ctx, err, "failed to write gzipped epg")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := ctx.Value(constant.ContextStreamData).(*url_generator.Data)

	logging.Debug(ctx, "handling proxy request", "type", data.RequestType)

	switch data.RequestType {
	case url_generator.File:
		s.handleFileProxy(ctx, w, data)
	case url_generator.Stream:
		s.handleStreamProxy(ctx, w, r)
	default:
		logging.Error(ctx, nil, "invalid proxy request type", "type", data.RequestType)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (s *Server) handleFileProxy(ctx context.Context, w http.ResponseWriter, data *url_generator.Data) {
	logging.Debug(ctx, "proxying file", "url", data.URL)

	reader, err := s.cache.NewReader(ctx, data.URL)
	if err != nil {
		logging.Error(ctx, err, "file proxy failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	contentType := reader.GetContentType()
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)

	if _, err = io.Copy(w, reader); err != nil {
		logging.Error(ctx, err, "file copy failed")
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (s *Server) prepareEPGStreamer(ctx context.Context) (*xmltv.Streamer, error) {
	c := ctx.Value(constant.ContextClient).(*manager.Client)

	m3u8Streamer := m3u8.NewStreamer(c.GetSubscriptions(), "", s.cache)
	channels, err := m3u8Streamer.GetAllChannels(ctx)
	if err != nil {
		logging.Error(ctx, err, "failed to get channels")
		return nil, err
	}

	return xmltv.NewStreamer(c.GetSubscriptions(), s.cache, channels), nil
}

func generateHash(parts ...any) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(fmt.Sprint(part)))
	}
	return hex.EncodeToString(h.Sum(nil)[:4])
}

func (s *Server) handleStreamProxy(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logging.Debug(ctx, "proxying stream")

	sub := ctx.Value(constant.ContextSubscription).(*manager.Subscription)
	data := ctx.Value(constant.ContextStreamData).(*url_generator.Data)

	var streamKey string
	url := data.URL
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
		streamKey = generateHash(url, time.Now().Unix())
	} else {
		streamKey = generateHash(data.URL)
	}

	if time.Since(data.Created) > linkTTL {
		logging.Error(ctx, errors.New("stream url expired"), "")
		sub.ExpiredCommandStreamer().Stream(ctx, w)
		return
	}

	if !s.acquireSemaphores(ctx) {
		logging.Error(ctx, errors.New("failed to acquire semaphores"), "")
		sub.LimitStreamer().Stream(ctx, w)
		return
	}
	defer s.releaseSemaphores(ctx)

	streamSource := sub.LinkStreamer(url)
	if streamSource == nil {
		logging.Error(ctx, errors.New("failed to create stream source"), "")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	streamReq := video.StreamRequest{
		StreamKey:  streamKey,
		StreamData: streamSource,
		Context:    ctx,
		Semaphore:  sub.GetSemaphore(),
	}

	reader, err := s.streamManager.GetReader(streamReq)
	if errors.Is(err, video.ErrSubscriptionSemaphore) {
		logging.Error(ctx, err, "failed to get stream")
		sub.LimitStreamer().Stream(ctx, w)
		return
	}
	if err != nil {
		logging.Error(ctx, err, "failed to get stream")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", streamContentType)
	written, err := io.Copy(w, reader)

	if err == nil && written == 0 {
		logging.Error(ctx, errors.New("no data written to response"), "")
		sub.UpstreamErrorStreamer().Stream(ctx, w)
	}

	if err != nil && !errors.Is(err, io.ErrClosedPipe) {
		logging.Error(ctx, err, "error copying stream to response")
	}
}
