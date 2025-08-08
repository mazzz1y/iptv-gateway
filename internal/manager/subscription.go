package manager

import (
	"fmt"
	"golang.org/x/sync/semaphore"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/shell"
	"iptv-gateway/internal/url_generator"
	"net/url"
)

type URLGenerator interface {
	CreateURL(data url_generator.Data) (*url.URL, error)
	Decrypt(s string) (*url_generator.Data, error)
}

type Subscription struct {
	name      string
	playlists []string
	epgs      []string

	urlGenerator  URLGenerator
	semaphore     *semaphore.Weighted
	proxyConfig   config.Proxy
	excludeConfig config.Excludes

	linkStreamer          *shell.CommandBuilder
	rateLimitStreamer     *shell.CommandBuilder
	upstreamErrorStreamer *shell.CommandBuilder
	expiredLinkStreamer   *shell.CommandBuilder
}

func NewSubscription(
	name string, urlGen URLGenerator, playlists []string, epgs []string,
	proxy config.Proxy, excludes config.Excludes, sem *semaphore.Weighted) (*Subscription, error) {

	streamStreamer, err := shell.NewCommandBuilder(
		proxy.Stream.Command,
		proxy.Stream.EnvVars,
		proxy.Stream.TemplateVars,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream command: %w", err)
	}

	rateLimitStreamer, err := shell.NewCommandBuilder(
		proxy.Error.RateLimitExceeded.Command,
		proxy.Error.RateLimitExceeded.EnvVars,
		proxy.Error.RateLimitExceeded.TemplateVars,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limit command: %w", err)
	}

	upstreamErrorStreamer, err := shell.NewCommandBuilder(
		proxy.Error.UpstreamError.Command,
		proxy.Error.UpstreamError.EnvVars,
		proxy.Error.UpstreamError.TemplateVars,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream error command: %w", err)
	}

	expiredLinkStreamer, err := shell.NewCommandBuilder(
		proxy.Error.LinkExpired.Command,
		proxy.Error.LinkExpired.EnvVars,
		proxy.Error.LinkExpired.TemplateVars,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expired link command: %w", err)
	}

	return &Subscription{
		name:                  name,
		urlGenerator:          urlGen,
		playlists:             playlists,
		epgs:                  epgs,
		semaphore:             sem,
		proxyConfig:           proxy,
		excludeConfig:         excludes,
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

func (s *Subscription) GetURLGenerator() URLGenerator {
	return s.urlGenerator
}
func (s *Subscription) GetExcludes() config.Excludes {
	return s.excludeConfig
}

func (s *Subscription) GetSemaphore() *semaphore.Weighted {
	return s.semaphore
}

func (s *Subscription) IsProxied() bool {
	return s.proxyConfig.Enabled != nil && *s.proxyConfig.Enabled
}

func (s *Subscription) LinkStreamer(streamUrl string) *shell.CommandBuilder {
	return s.linkStreamer.WithTemplateVars(map[string]any{"url": streamUrl})
}

func (s *Subscription) LimitStreamer() *shell.CommandBuilder {
	return s.rateLimitStreamer
}

func (s *Subscription) UpstreamErrorStreamer() *shell.CommandBuilder {
	return s.upstreamErrorStreamer
}

func (s *Subscription) ExpiredCommandStreamer() *shell.CommandBuilder {
	return s.expiredLinkStreamer
}
