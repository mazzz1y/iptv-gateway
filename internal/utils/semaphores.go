package utils

import (
	"context"
	"errors"
	"iptv-gateway/internal/ctxutil"
	"iptv-gateway/internal/logging"
	"time"

	"golang.org/x/sync/semaphore"
)

const (
	SemaphoreTimeout = 3 * time.Second
)

func AcquireSemaphore(ctx context.Context, sem *semaphore.Weighted, name string) bool {
	if sem == nil {
		return true
	}

	semCtx, cancel := context.WithTimeout(ctx, SemaphoreTimeout)
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

func TryAcquireSemaphore(ctx context.Context, sem *semaphore.Weighted, name string) bool {
	if sem == nil {
		return true
	}

	ctx = ctxutil.WithSemaphoreName(ctx, name)

	logging.Debug(ctx, "trying to acquire semaphore")

	if ok := sem.TryAcquire(1); !ok {
		logging.Debug(ctx, "semaphore not available")
		return false
	}

	logging.Debug(ctx, "semaphore acquired")
	return true
}
