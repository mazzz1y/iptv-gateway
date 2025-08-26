package server

import (
	"context"
	"errors"
	"iptv-gateway/internal/app"
	"iptv-gateway/internal/cache"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/demux"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/metrics"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	muxSecretVar        = "secret"
	muxEncryptedDataVar = "encrypted_data"
)

type Server struct {
	router  *mux.Router
	server  *http.Server
	manager *app.Manager

	cache      *cache.Cache
	httpClient *http.Client

	demux *demux.Demuxer

	serverURL     string
	listenAddr    string
	metricsServer *http.Server

	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(cfg *config.Config) (*Server, error) {
	m, err := app.NewManager(cfg)
	if err != nil {
		return nil, err
	}

	c, err := cache.NewCache(cfg.Cache.Path,
		time.Duration(cfg.Cache.TTL), time.Duration(cfg.Cache.Retention))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		router:     mux.NewRouter(),
		manager:    m,
		cache:      c,
		httpClient: c.NewCachedHTTPClient(),
		demux:      demux.NewDemuxer(),
		serverURL:  cfg.PublicURL.String(),
		listenAddr: cfg.ListenAddr,
		ctx:        ctx,
		cancel:     cancel,
	}

	if cfg.MetricsAddr != "" {
		server.setupMetricsServer(cfg.MetricsAddr)
	}

	return server, nil
}

func (s *Server) Start() error {
	s.setupRoutes()

	if s.metricsServer != nil {
		go func() {
			logging.Info(s.ctx, "starting metrics server", "address", s.metricsServer.Addr)
			if err := s.metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logging.Error(s.ctx, err, "metrics server failed")
			}
		}()
	}

	s.server = &http.Server{
		Addr:    s.listenAddr,
		Handler: s.router,
	}

	logging.Info(s.ctx, "starting http server", "address", s.listenAddr)

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logging.Error(s.ctx, err, "server failed")
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	s.cancel()

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	s.cache.Close()
	s.demux.Stop()

	logging.Info(ctx, "stopping http server")
	if err := s.server.Shutdown(ctx); err != nil {
		logging.Error(ctx, err, "server shutdown timeout, force closing connections")
		s.server.Close()
	}

	if s.metricsServer != nil {
		logging.Info(ctx, "stopping metrics server")
		if err := s.metricsServer.Shutdown(ctx); err != nil {
			logging.Error(ctx, err, "metrics server shutdown timeout, force closing connections")
			s.metricsServer.Close()
		}
	}

	logging.Info(ctx, "server stopped")
	return nil
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/healthz", s.handleHealthz)

	if s.metricsServer == nil {
		s.router.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))
	}

	authRouter := s.router.PathPrefix("/{" + muxSecretVar + "}").Subrouter()

	authRouter.Use(s.requestIDMiddleware)
	authRouter.Use(s.loggerMiddleware)
	authRouter.Use(s.authenticationMiddleware)

	authRouter.HandleFunc("/playlist.m3u8", s.handlePlaylist)
	authRouter.HandleFunc("/epg.xml", s.handleEPG)
	authRouter.HandleFunc("/epg.xml.gz", s.handleEPGgz)

	proxyRouter := authRouter.PathPrefix("/{" + muxEncryptedDataVar + "}").Subrouter()
	proxyRouter.Use(s.decryptProxyDataMiddleware)
	proxyRouter.HandleFunc("/{.*}", s.handleProxy)

	authRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}

func (s *Server) setupMetricsServer(addr string) {
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))

	s.metricsServer = &http.Server{
		Addr:    addr,
		Handler: m,
	}
}
