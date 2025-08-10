package utils

import (
	"context"
	"errors"
	"golang.org/x/sync/semaphore"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/logging"
)

func AcquireSemaphore(ctx context.Context, sem *semaphore.Weighted, name string) bool {
	if sem == nil {
		return true
	}

	semCtx, cancel := context.WithTimeout(ctx, constant.SemaphoreTimeout)
	defer cancel()
	semCtx = context.WithValue(semCtx, constant.ContextSemaphoreName, name)

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
