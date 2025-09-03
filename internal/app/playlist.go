package app

import (
	"fmt"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/shell"
	"iptv-gateway/internal/urlgen"

	"golang.org/x/sync/semaphore"
)

type PlaylistSubscription struct {
	name string

	sources []string

	urlGenerator *urlgen.Generator
	semaphore    *semaphore.Weighted

	channelRules  []rules.ChannelRule
	playlistRules []rules.PlaylistRule

	proxyConfig config.Proxy

	linkStreamer          *shell.Streamer
	rateLimitStreamer     *shell.Streamer
	upstreamErrorStreamer *shell.Streamer
	expiredLinkStreamer   *shell.Streamer
}

func NewPlaylistSubscription(
	name string, urlGen urlgen.Generator,
	sources []string,
	proxy config.Proxy, channelRules []rules.ChannelRule, playlistRules []rules.PlaylistRule,
	sem *semaphore.Weighted) (*PlaylistSubscription, error) {

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

	return &PlaylistSubscription{
		name:                  name,
		urlGenerator:          &urlGen,
		sources:               sources,
		semaphore:             sem,
		proxyConfig:           proxy,
		channelRules:          channelRules,
		playlistRules:         playlistRules,
		linkStreamer:          streamStreamer,
		rateLimitStreamer:     rateLimitStreamer,
		upstreamErrorStreamer: upstreamErrorStreamer,
		expiredLinkStreamer:   expiredLinkStreamer,
	}, nil
}

func (ps *PlaylistSubscription) Name() string {
	return ps.name
}

func (ps *PlaylistSubscription) Playlists() []string {
	return ps.sources
}

func (ps *PlaylistSubscription) URLGenerator() *urlgen.Generator {
	return ps.urlGenerator
}

func (ps *PlaylistSubscription) ChannelRules() []rules.ChannelRule {
	return ps.channelRules
}

func (ps *PlaylistSubscription) PlaylistRules() []rules.PlaylistRule {
	return ps.playlistRules
}

func (ps *PlaylistSubscription) Semaphore() *semaphore.Weighted {
	return ps.semaphore
}

func (ps *PlaylistSubscription) IsProxied() bool {
	return ps.proxyConfig.Enabled != nil && *ps.proxyConfig.Enabled
}

func (ps *PlaylistSubscription) ProxyConfig() config.Proxy {
	return ps.proxyConfig
}

func (ps *PlaylistSubscription) LinkStreamer(streamUrl string) *shell.Streamer {
	return ps.linkStreamer.WithTemplateVars(map[string]any{"url": streamUrl})
}

func (ps *PlaylistSubscription) LimitStreamer() *shell.Streamer {
	return ps.rateLimitStreamer
}

func (ps *PlaylistSubscription) UpstreamErrorStreamer() *shell.Streamer {
	return ps.upstreamErrorStreamer
}

func (ps *PlaylistSubscription) ExpiredCommandStreamer() *shell.Streamer {
	return ps.expiredLinkStreamer
}


