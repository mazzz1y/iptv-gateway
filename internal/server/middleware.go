package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/gorilla/mux"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/manager"
	"iptv-gateway/internal/streamer/video"
	"iptv-gateway/internal/url_generator"
	"net/http"
	"time"
)

func (s *Server) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(startTime)
		logging.HttpRequest(r.Context(), r, rw.statusCode, duration)
	})
}

func (s *Server) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 4)
		rand.Read(b)
		ctx := context.WithValue(r.Context(), constant.ContextRequestID, hex.EncodeToString(b))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		secret, ok := vars["secret"]
		if !ok || secret == "" {
			logging.Error(r.Context(), "authentication failed: no secret")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		c := s.manager.GetClient(secret)
		if c == nil {
			logging.Error(r.Context(), "authentication failed: invalid secret", "secret", secret)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), constant.ContextClient, c)
		ctx = context.WithValue(ctx, constant.ContextClientName, c.GetName())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) decryptProxyDataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		pp, ok := vars["encrypted_data"]
		if !ok {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		c, _ := r.Context().Value(constant.ContextClient).(*manager.Client)

		for _, sub := range c.GetSubscriptions() {
			urlGen := sub.GetURLGenerator()
			if urlGen != nil {
				data, err := urlGen.Decrypt(pp)
				if err == nil {
					ctx := context.WithValue(r.Context(), constant.ContextSubscription, sub)
					ctx = context.WithValue(ctx, constant.ContextSubscriptionName, sub.GetName())
					ctx = context.WithValue(ctx, constant.ContextStreamData, data)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				if errors.Is(err, url_generator.ErrExpiredURL) {
					ctx := context.WithValue(r.Context(), constant.ContextSubscription, sub)
					ctx = context.WithValue(ctx, constant.ContextSubscriptionName, sub.GetName())
					streamer := video.NewStreamer(sub.ExpiredCommand())
					streamer.Stream(ctx, w)
					return
				}
			}
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}
