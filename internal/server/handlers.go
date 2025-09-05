package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"iptv-gateway/internal/app"
	"iptv-gateway/internal/ctxutil"
	"iptv-gateway/internal/demux"
	"iptv-gateway/internal/listing/m3u8"
	"iptv-gateway/internal/listing/xmltv"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/metrics"
	"iptv-gateway/internal/urlgen"
)

const (
	streamContentType = "video/mp2t"
)

func (s *Server) handlePlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = ctxutil.WithRequestType(ctx, metrics.RequestTypePlaylist)

	c := ctxutil.Client(ctx).(*app.Client)

	logging.Debug(ctx, "playlist request")

	w.Header().Set("Content-Type", "application/x-mpegurl")
	w.Header().Set("Cache-Control", "no-cache")

	streamer := m3u8.NewStreamer(c.PlaylistSubscriptions(), c.EpgLink(), s.httpClient)
	if _, err := streamer.WriteTo(ctx, w); err != nil {
		logging.Error(ctx, err, "failed to write playlist")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	metrics.ListingDownloadTotal.WithLabelValues(c.Name(), metrics.RequestTypePlaylist).Inc()
}

func (s *Server) handleEPG(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = ctxutil.WithRequestType(ctx, metrics.RequestTypeEPG)

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

	c := ctxutil.Client(ctx).(*app.Client)
	metrics.ListingDownloadTotal.WithLabelValues(c.Name(), metrics.RequestTypeEPG).Inc()
}

func (s *Server) handleEPGgz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = ctxutil.WithRequestType(ctx, metrics.RequestTypeEPG)

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

	c := ctxutil.Client(ctx).(*app.Client)
	metrics.ListingDownloadTotal.WithLabelValues(c.Name(), metrics.RequestTypeEPG).Inc()
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := ctxutil.StreamData(ctx).(*urlgen.Data)

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
	ctx = ctxutil.WithRequestType(ctx, metrics.RequestTypeFile)

	req, err := http.NewRequestWithContext(ctx, "GET", data.URL, nil)
	if err != nil {
		logging.Error(ctx, err, "failed to create request")
		http.Error(w, "Failed to create request", http.StatusBadGateway)
		return
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logging.Error(ctx, err, "file proxy failed")
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		logging.Error(ctx, fmt.Errorf("status code: %d", resp.StatusCode), "upstream returned error")
		http.Error(w, http.StatusText(resp.StatusCode), resp.StatusCode)
		return
	}

	for header, values := range resp.Header {
		w.Header()[header] = values
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
	c := ctxutil.Client(ctx).(*app.Client)

	m3u8Streamer := m3u8.NewStreamer(c.PlaylistSubscriptions(), "", s.httpClient)
	channels, err := m3u8Streamer.GetAllChannels(ctx)
	if err != nil {
		logging.Error(ctx, err, "failed to get channels")
		return nil, err
	}

	return xmltv.NewStreamer(c.EPGSubscriptions(), s.httpClient, channels), nil
}

func (s *Server) handleStreamProxy(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logging.Debug(ctx, "proxying stream")

	playlist := ctxutil.Provider(ctx).(*app.PlaylistSubscription)
	data := ctxutil.StreamData(ctx).(*urlgen.Data)
	ctx = ctxutil.WithChannelID(ctx, data.ChannelID)
	ctx = ctxutil.WithChannelHidden(ctx, data.Hidden)

	clientName := ctxutil.ClientName(ctx)
	playlistName := playlist.Name()

	streamKey := generateStreamKey(data.URL, r.URL.RawQuery)

	if !s.acquireSemaphores(ctx) {
		logging.Error(ctx, errors.New("failed to acquire semaphores"), "")
		playlist.LimitStreamer().Stream(ctx, w)
		return
	}
	defer s.releaseSemaphores(ctx)

	url := data.URL
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}

	streamSource := playlist.LinkStreamer(url)
	if streamSource == nil {
		logging.Error(ctx, errors.New("failed to create stream source"), "")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	demuxReq := demux.Request{
		Context:   ctx,
		StreamKey: streamKey,
		Streamer:  streamSource,
		Semaphore: playlist.Semaphore(),
	}

	reader, err := s.demux.GetReader(demuxReq)
	if errors.Is(err, demux.ErrSubscriptionSemaphore) {
		logging.Error(ctx, err, "failed to get stream")
		if !data.Hidden {
			metrics.StreamsFailuresTotal.WithLabelValues(
				clientName, playlist.Name(), data.ChannelID, metrics.FailureReasonPlaylistLimit).Inc()
		}
		playlist.LimitStreamer().Stream(ctx, w)
		return
	}
	if err != nil {
		logging.Error(ctx, err, "failed to get stream")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	if !data.Hidden {
		metrics.ClientStreamsActive.WithLabelValues(clientName, playlistName, data.ChannelID).Inc()
		defer metrics.ClientStreamsActive.WithLabelValues(clientName, playlistName, data.ChannelID).Dec()
	}

	w.Header().Set("Content-Type", streamContentType)
	written, err := io.Copy(w, reader)

	if err == nil && written == 0 {
		if !data.Hidden {
			logging.Error(ctx, errors.New("no data written to response"), "")
			metrics.StreamsFailuresTotal.WithLabelValues(
				clientName, playlist.Name(), data.ChannelID, metrics.FailureReasonUpstreamError).Inc()
			playlist.UpstreamErrorStreamer().Stream(ctx, w)
		}
		return
	}

	if err != nil && !errors.Is(err, io.ErrClosedPipe) {
		logging.Error(ctx, err, "error copying stream to response")
	}
}

func generateStreamKey(url, query string) string {
	if query != "" {
		return generateHash(url+"?"+query, time.Now().Unix())
	}
	return generateHash(url)
}

func generateHash(parts ...any) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(fmt.Sprint(part)))
	}
	return hex.EncodeToString(h.Sum(nil)[:4])
}
