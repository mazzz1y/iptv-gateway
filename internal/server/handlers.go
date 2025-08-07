package server

import (
	"context"
	"io"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/manager"
	"iptv-gateway/internal/streamer/m3u8"
	"iptv-gateway/internal/streamer/xmltv"
	"iptv-gateway/internal/url_generator"
	"net/http"
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
		s.handleStreamProxy(ctx, w, r, data)
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

func (s *Server) handleStreamProxy(ctx context.Context, w http.ResponseWriter, r *http.Request, data *url_generator.Data) {
	sub := ctx.Value(constant.ContextSubscription).(*manager.Subscription)

	if !s.acquireSemaphores(ctx) {
		sub.LimitStreamer().Stream(ctx, w)
		return
	}
	defer s.releaseSemaphores(ctx)

	urlWithParams := data.URL
	if r.URL.RawQuery != "" {
		urlWithParams = data.URL + "?" + r.URL.RawQuery
	}

	logging.Info(ctx, "streaming", "channel", data.ChannelID)

	streamer := sub.LinkStreamer(urlWithParams)
	w.Header().Set("Content-Type", streamer.ContentType())

	bWritten, err := streamer.Stream(ctx, w)
	if err != nil {
		logging.Error(ctx, err, "video stream failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if bWritten == 0 {
		if _, err := sub.UpstreamErrorStreamer().Stream(ctx, w); err != nil {
			logging.Error(ctx, err, "fallback stream failed")
		}
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
