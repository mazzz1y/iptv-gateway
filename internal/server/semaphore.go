package server

import (
	"context"
	"fmt"
	"iptv-gateway/internal/app"
	"iptv-gateway/internal/ctxutil"
	"iptv-gateway/internal/metrics"
	"iptv-gateway/internal/urlgen"
	"iptv-gateway/internal/utils"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func (s *Server) acquireSemaphores(ctx context.Context) bool {
	c := ctxutil.Client(ctx).(*app.Client)
	data := ctxutil.StreamData(ctx).(*urlgen.Data)

	g, gCtx := errgroup.WithContext(ctx)

	acquireSem := func(sem *semaphore.Weighted, reason string) func() error {
		return func() error {
			if sem == nil || utils.AcquireSemaphore(gCtx, sem, reason) {
				return nil
			}
			metrics.StreamsFailuresTotal.WithLabelValues(
				c.Name(), ctxutil.ProviderName(ctx), data.ChannelID, reason).Inc()
			return fmt.Errorf("failed to acquire semaphore: %s", reason)
		}
	}

	if managerSem := s.manager.GlobalSemaphore(); managerSem != nil {
		g.Go(acquireSem(managerSem, metrics.FailureReasonGlobalLimit))
	}

	if clientSem := c.Semaphore(); clientSem != nil {
		g.Go(acquireSem(clientSem, metrics.FailureReasonClientLimit))
	}

	return g.Wait() == nil
}

func (s *Server) releaseSemaphores(ctx context.Context) {
	c := ctxutil.Client(ctx).(*app.Client)

	if managerSem := s.manager.GlobalSemaphore(); managerSem != nil {
		managerSem.Release(1)
	}

	if clientSem := c.Semaphore(); clientSem != nil {
		clientSem.Release(1)
	}
}
