package server

import (
	"errors"
	"iptv-gateway/internal/ctxutil"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/urlgen"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (s *Server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(startTime)
		logging.HttpRequest(r.Context(), r, rw.statusCode, duration, rw.bytesWritten)
	})
}

func (s *Server) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ctxutil.WithRequestID(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) clientAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		secret, ok := vars[muxClientSecretVar]
		if !ok || secret == "" {
			logging.Debug(r.Context(), "authentication failed: no secret")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		client := s.manager.Client(secret)
		if client == nil {
			logging.Debug(r.Context(), "authentication failed: invalid secret", "secret", secret)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		ctx := ctxutil.WithClient(r.Context(), client)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) proxyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		token, ok := vars[muxEncryptedTokenVar]
		if !ok || token == "" {
			logging.Debug(r.Context(), "authentication failed: no token")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		ctx := r.Context()
		for _, c := range s.manager.Clients() {
			for _, pr := range c.Providers() {
				data, err := pr.URLGenerator().Decrypt(token)
				if err == nil {
					ctx = ctxutil.WithClient(ctx, c)
					ctx = ctxutil.WithProvider(ctx, pr)
					ctx = ctxutil.WithStreamData(ctx, data)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}

				if errors.Is(err, urlgen.ErrExpiredStreamURL) {
					if pr.ExpiredLinkStreamer() != nil {
						ctx = ctxutil.WithClient(ctx, c)
						ctx = ctxutil.WithProvider(ctx, pr)
						pr.ExpiredLinkStreamer().Stream(ctx, w)
					} else {
						http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					}
					return
				}
			}
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}
