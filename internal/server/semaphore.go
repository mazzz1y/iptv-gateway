package server

import (
	"context"
	"errors"
	"iptv-gateway/internal/client"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/utils"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func (s *Server) acquireSemaphores(ctx context.Context) bool {
	c, _ := ctx.Value(constant.ContextClient).(*client.Client)

	g, gCtx := errgroup.WithContext(ctx)

	acquireSem := func(sem *semaphore.Weighted, semType string) func() error {
		return func() error {
			if sem == nil || utils.AcquireSemaphore(gCtx, sem, semType) {
				return nil
			}
			return errors.New("failed to acquire " + semType + " semaphore")
		}
	}

	if managerSem := s.manager.GetSemaphore(); managerSem != nil {
		g.Go(acquireSem(managerSem, "global"))
	}

	if clientSem := c.GetSemaphore(); clientSem != nil {
		g.Go(acquireSem(clientSem, "client"))
	}

	return g.Wait() == nil
}

func (s *Server) releaseSemaphores(ctx context.Context) {
	c, _ := ctx.Value(constant.ContextClient).(*client.Client)

	if managerSem := s.manager.GetSemaphore(); managerSem != nil {
		managerSem.Release(1)
	}

	if clientSem := c.GetSemaphore(); clientSem != nil {
		clientSem.Release(1)
	}
}
