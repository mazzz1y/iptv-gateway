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

		for pr := range client.URLProviders() {
			if s.tryDecryptWithURLProvider(r, w, next, token, pr) {
				return
			}
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func (s *Server) tryDecryptWithURLProvider(r *http.Request, w http.ResponseWriter,
	next http.Handler, token string, pr app.ProviderWithURLGen) bool {
	data, err := pr.URLGen.Decrypt(token)
	if err == nil {
		ctx := ctxutil.WithProvider(r.Context(), pr.Provider)
		ctx = ctxutil.WithStreamData(ctx, data)
		next.ServeHTTP(w, r.WithContext(ctx))
		return true
	}

	if errors.Is(err, urlgen.ErrExpiredStreamURL) {
		ctx := ctxutil.WithProvider(r.Context(), pr.Provider)
		if pr.ExpiredStreamer != nil {
			pr.ExpiredStreamer.Stream(ctx, w)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		return true
	}

	return false
}
