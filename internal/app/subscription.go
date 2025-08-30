package app

import (
	"fmt"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/shell"
	"iptv-gateway/internal/urlgen"

	"golang.org/x/sync/semaphore"
)

type Subscription struct {
	name      string
	playlists []string
	epgs      []string

	urlGenerator *urlgen.Generator
	semaphore    *semaphore.Weighted
	rules        []rules.RuleAction

	proxyConfig config.Proxy

	linkStreamer          *shell.Streamer
	rateLimitStreamer     *shell.Streamer
	upstreamErrorStreamer *shell.Streamer
	expiredLinkStreamer   *shell.Streamer
}

func NewSubscription(
	name string, urlGen urlgen.Generator, playlists []string, epgs []string,
	proxy config.Proxy, r []rules.RuleAction, sem *semaphore.Weighted) (*Subscription, error) {

	streamStreamer, err := shell.NewShellStreamer(
		proxy.Stream.Command,
		proxy.Stream.EnvVars,
		proxy.Stream.TemplateVars,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream command: %w", err)
	}

	rateLimitStreamer, err := shell.NewShellStreamer(
		proxy.Error.RateLimitExceeded.Command,
		proxy.Error.RateLimitExceeded.EnvVars,
		proxy.Error.RateLimitExceeded.TemplateVars,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limit command: %w", err)
	}

	upstreamErrorStreamer, err := shell.NewShellStreamer(
		proxy.Error.UpstreamError.Command,
		proxy.Error.UpstreamError.EnvVars,
		proxy.Error.UpstreamError.TemplateVars,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream error command: %w", err)
	}

	expiredLinkStreamer, err := shell.NewShellStreamer(
		proxy.Error.LinkExpired.Command,
		proxy.Error.LinkExpired.EnvVars,
		proxy.Error.LinkExpired.TemplateVars,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expired link command: %w", err)
	}

	return &Subscription{
		name:                  name,
		urlGenerator:          &urlGen,
		playlists:             playlists,
		epgs:                  epgs,
		semaphore:             sem,
		proxyConfig:           proxy,
		rules:                 r,
		linkStreamer:          streamStreamer,
		rateLimitStreamer:     rateLimitStreamer,
		upstreamErrorStreamer: upstreamErrorStreamer,
		expiredLinkStreamer:   expiredLinkStreamer,
	}, nil
}

func (s *Subscription) GetName() string {
	return s.name
}

func (s *Subscription) GetPlaylists() []string {
	return s.playlists
}

func (s *Subscription) GetEPGs() []string {
	return s.epgs
}

func (s *Subscription) GetURLGenerator() *urlgen.Generator {
	return s.urlGenerator
}
func (s *Subscription) GetRules() []rules.RuleAction {
	return s.rules
}

func (s *Subscription) GetSemaphore() *semaphore.Weighted {
	return s.semaphore
}

func (s *Subscription) IsProxied() bool {
	return s.proxyConfig.Enabled != nil && *s.proxyConfig.Enabled
}

func (s *Subscription) LinkStreamer(streamUrl string) *shell.Streamer {
	return s.linkStreamer.WithTemplateVars(map[string]any{"url": streamUrl})
}

func (s *Subscription) LimitStreamer() *shell.Streamer {
	return s.rateLimitStreamer
}

func (s *Subscription) UpstreamErrorStreamer() *shell.Streamer {
	return s.upstreamErrorStreamer
}

func (s *Subscription) ExpiredCommandStreamer() *shell.Streamer {
	return s.expiredLinkStreamer
}
