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
		token, ok := vars[muxEncryptedDataVar]
		if !ok {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		client := ctxutil.Client(r.Context()).(*app.Client)

		for sub := range client.URLGenerators() {
			if s.tryDecryptWithSubscription(r, w, next, token, sub) {
				return
			}
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func (s *Server) tryDecryptWithSubscription(r *http.Request, w http.ResponseWriter,
	next http.Handler, token string, sub app.URLGeneratorSubscription) bool {
	data, err := sub.URLGen.Decrypt(token)
	if err == nil {
		ctx := ctxutil.WithSubscription(r.Context(), sub.Subscription)
		ctx = ctxutil.WithStreamData(ctx, data)
		next.ServeHTTP(w, r.WithContext(ctx))
		return true
	}

	if errors.Is(err, urlgen.ErrExpiredStreamURL) {
		ctx := ctxutil.WithSubscription(r.Context(), sub.Subscription)
		if sub.ExpiredStreamer != nil {
			sub.ExpiredStreamer.Stream(ctx, w)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		return true
	}

	return false
}
