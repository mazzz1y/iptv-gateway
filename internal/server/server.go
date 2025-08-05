package server

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"iptv-gateway/internal/cache"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/manager"
	"net/http"
	"time"
)

type Server struct {
	router  *mux.Router
	server  *http.Server
	manager *manager.Manager
	cache   *cache.Cache

	serverURL  string
	listenAddr string
}

func NewServer(cfg *config.Config) (*Server, error) {
	m, err := manager.NewManager(cfg)
	if err != nil {
		logging.Error(context.TODO(), "failed to initialize manager", "error", err)
		return nil, err
	}

	c, err := cache.NewCache(cfg.Cache.Path, time.Duration(cfg.Cache.TTL))
	if err != nil {
		logging.Error(context.TODO(), "failed to create cache", "error", err)
		return nil, err
	}

	server := &Server{
		router:     mux.NewRouter(),
		manager:    m,
		cache:      c,
		serverURL:  cfg.PublicURL.String(),
		listenAddr: cfg.ListenAddr,
	}

	return server, nil
}

func (s *Server) Start() error {
	s.setupRoutes()

	s.server = &http.Server{
		Addr:    s.listenAddr,
		Handler: s.router,
	}

	logging.Info(context.TODO(), "starting http server", "address", s.listenAddr)

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logging.Error(context.TODO(), "server failed", "error", err)
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.cache.Close()

	logging.Info(ctx, "stopping http server")

	if err := s.server.Shutdown(ctx); err != nil {
		logging.Error(ctx, "server shutdown failed", "error", err)
		return err
	}

	logging.Info(ctx, "http server stopped")
	return nil
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/healthz", s.handleHealthz)

	authRouter := s.router.PathPrefix("/{secret}").Subrouter()

	authRouter.Use(s.requestIDMiddleware)
	authRouter.Use(s.loggerMiddleware)
	authRouter.Use(s.authenticationMiddleware)

	authRouter.HandleFunc("/playlist.m3u8", s.handlePlaylist)
	authRouter.HandleFunc("/epg.xml", s.handleEPG)
	authRouter.HandleFunc("/epg.xml.gz", s.handleEPGgz)

	proxyRouter := authRouter.PathPrefix("/{encrypted_data}").Subrouter()
	proxyRouter.Use(s.decryptProxyDataMiddleware)
	proxyRouter.HandleFunc("/{.*}", s.handleProxy)

	authRouter.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}
