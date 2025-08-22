package server

import (
	"errors"
	"iptv-gateway/internal/app"
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

func (s *Server) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		secret, ok := vars[muxSecretVar]
		if !ok || secret == "" {
			logging.Error(r.Context(), nil, "authentication failed: no secret")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		c := s.manager.GetClient(secret)
		if c == nil {
			logging.Error(r.Context(), nil, "authentication failed: invalid secret", "secret", secret)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		ctx := ctxutil.WithClient(r.Context(), c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) decryptProxyDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pp, ok := vars[muxEncryptedDataVar]
		if !ok {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		c := ctxutil.Client(r.Context()).(*app.Client)
		for _, sub := range c.GetSubscriptions() {
			urlGen := sub.GetURLGenerator()
			if urlGen != nil {
				data, err := urlGen.Decrypt(pp)
				if err == nil {
					ctx := ctxutil.WithSubscription(r.Context(), sub)
					ctx = ctxutil.WithStreamData(ctx, data)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				if errors.Is(err, urlgen.ErrExpiredStreamURL) {
					ctx := ctxutil.WithSubscription(r.Context(), sub)
					sub.ExpiredCommandStreamer().Stream(ctx, w)
					return
				}
			}
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}
