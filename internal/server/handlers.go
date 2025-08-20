package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"iptv-gateway/internal/client"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/demux"
	"iptv-gateway/internal/listing/m3u8"
	"iptv-gateway/internal/listing/xmltv"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/urlgen"
	"net/http"
	"time"
)

const (
	streamContentType = "video/mp2t"
)

func (s *Server) handlePlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c := ctx.Value(constant.ContextClient).(*client.Client)

	logging.Debug(ctx, "playlist request")

	w.Header().Set("Content-Type", "application/x-mpegurl")
	w.Header().Set("Cache-Control", "no-cache")

	streamer := m3u8.NewStreamer(c.GetSubscriptions(), c.GetEpgLink(), s.cache.NewCachedHTTPClient())
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
	data := ctx.Value(constant.ContextStreamData).(*urlgen.Data)

	logging.Debug(ctx, "handling proxy request", "type", data.RequestType)

	switch data.RequestType {
	case urlgen.File:
		s.handleFileProxy(ctx, w, data)
	case urlgen.Stream:
		s.handleStreamProxy(ctx, w, r)
	default:
		logging.Error(ctx, nil, "invalid proxy request type", "type", data.RequestType)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (s *Server) handleFileProxy(ctx context.Context, w http.ResponseWriter, data *urlgen.Data) {
	logging.Debug(ctx, "proxying file", "url", data.URL)

	c := s.cache.NewCachedHTTPClient()
	resp, err := c.Get(data.URL)
	if err != nil {
		logging.Error(ctx, err, "file proxy failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	if _, err = io.Copy(w, resp.Body); err != nil {
		logging.Error(ctx, err, "file copy failed")
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (s *Server) prepareEPGStreamer(ctx context.Context) (*xmltv.Streamer, error) {
	c := ctx.Value(constant.ContextClient).(*client.Client)

	m3u8Streamer := m3u8.NewStreamer(c.GetSubscriptions(), "", s.cache.NewCachedHTTPClient())
	channels, err := m3u8Streamer.GetAllChannels(ctx)
	if err != nil {
		logging.Error(ctx, err, "failed to get channels")
		return nil, err
	}

	return xmltv.NewStreamer(c.GetSubscriptions(), s.cache.NewCachedHTTPClient(), channels), nil
}

func (s *Server) handleStreamProxy(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logging.Debug(ctx, "proxying stream")

	subscription := ctx.Value(constant.ContextSubscription).(*client.Subscription)
	data := ctx.Value(constant.ContextStreamData).(*urlgen.Data)
	ctx = context.WithValue(ctx, constant.ContextChannelID, data.ChannelID)

	var streamKey string
	url := data.URL
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
		streamKey = generateHash(url, time.Now().Unix())
	} else {
		streamKey = generateHash(data.URL)
	}

	if !s.acquireSemaphores(ctx) {
		logging.Error(ctx, errors.New("failed to acquire semaphores"), "")
		subscription.LimitStreamer().Stream(ctx, w)
		return
	}
	defer s.releaseSemaphores(ctx)

	streamSource := subscription.LinkStreamer(url)
	if streamSource == nil {
		logging.Error(ctx, errors.New("failed to create stream source"), "")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	demuxReq := demux.Request{
		Context:   ctx,
		StreamKey: streamKey,
		Streamer:  streamSource,
		Semaphore: subscription.GetSemaphore(),
	}

	reader, err := s.demux.GetReader(demuxReq)
	if errors.Is(err, demux.ErrSubscriptionSemaphore) {
		logging.Error(ctx, err, "failed to get stream")
		subscription.LimitStreamer().Stream(ctx, w)
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
		subscription.UpstreamErrorStreamer().Stream(ctx, w)
	}

	if err != nil && !errors.Is(err, io.ErrClosedPipe) {
		logging.Error(ctx, err, "error copying stream to response")
	}
}

func generateHash(parts ...any) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(fmt.Sprint(part)))
	}
	return hex.EncodeToString(h.Sum(nil)[:4])
}
