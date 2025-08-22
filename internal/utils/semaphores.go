package utils

import (
	"context"
	"errors"
	"iptv-gateway/internal/ctxutil"
	"iptv-gateway/internal/logging"
	"time"

	"golang.org/x/sync/semaphore"
)

func AcquireSemaphore(ctx context.Context, sem *semaphore.Weighted, name string) bool {
	if sem == nil {
		return true
	}

	semCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	semCtx = ctxutil.WithSemaphoreName(semCtx, name)

	defer cancel()

	logging.Debug(ctx, "acquiring semaphore")

	if err := sem.Acquire(semCtx, 1); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logging.Info(ctx, "semaphore acquisition timeout")
		} else {
			logging.Error(ctx, err, "semaphore acquisition failed")
		}
		return false
	}

	logging.Debug(ctx, "semaphore acquired")
	return true
}
