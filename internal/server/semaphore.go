package server

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/manager"
	"time"
)

const acquireSemaphoreTimeout = 6 * time.Second

func (s *Server) acquireSemaphores(ctx context.Context) bool {
	c, _ := ctx.Value(constant.ContextClient).(*manager.Client)
	sub, _ := ctx.Value(constant.ContextSubscription).(*manager.Subscription)

	timeoutCtx, cancel := context.WithTimeout(ctx, acquireSemaphoreTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(timeoutCtx)

	acquireSem := func(sem *semaphore.Weighted, semType string) func() error {
		return func() error {
			if sem == nil || s.acquireWithTimeout(gCtx, sem, semType) {
				return nil
			}
			return errors.New("failed to acquire " + semType + " semaphore")
		}
	}

	if managerSem := s.manager.GetSemaphore(); managerSem != nil {
		g.Go(acquireSem(managerSem, "global"))
	}

	if subSem := sub.GetSemaphore(); subSem != nil {
		g.Go(acquireSem(subSem, "stream"))
	}

	if clientSem := c.GetSemaphore(); clientSem != nil {
		g.Go(acquireSem(clientSem, "manager"))
	}

	return g.Wait() == nil
}

func (s *Server) acquireWithTimeout(ctx context.Context, sem *semaphore.Weighted, semType string) bool {
	logging.Debug(ctx, "acquiring semaphore", "type", semType)

	if err := sem.Acquire(ctx, 1); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logging.Info(ctx, "semaphore acquisition timeout", "type", semType)
		} else {
			logging.Error(ctx, err, "semaphore acquisition failed", "type", semType)
		}
		return false
	}

	logging.Debug(ctx, "semaphore acquired", "type", semType)
	return true
}

func (s *Server) releaseSemaphores(ctx context.Context) {
	c, _ := ctx.Value(constant.ContextClient).(*manager.Client)
	sub, _ := ctx.Value(constant.ContextSubscription).(*manager.Subscription)

	if subSem := sub.GetSemaphore(); subSem != nil {
		subSem.Release(1)
	}

	if clientSem := c.GetSemaphore(); clientSem != nil {
		clientSem.Release(1)
	}

	if managerSem := s.manager.GetSemaphore(); managerSem != nil {
		managerSem.Release(1)
	}
}
