package manager

import (
	"golang.org/x/sync/semaphore"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/streamer/video"
	"iptv-gateway/internal/url_generator"
	"net/url"
)

type URLGenerator interface {
	CreateURL(data url_generator.Data) (*url.URL, error)
	Decrypt(s string) (*url_generator.Data, error)
}

type Subscription struct {
	name          string
	urlGenerator  URLGenerator
	playlists     []string
	epgs          []string
	semaphore     *semaphore.Weighted
	proxyConfig   config.Proxy
	excludeConfig config.Excludes
}

func NewSubscription(
	name string,
	urlGen URLGenerator,
	playlists []string,
	epgs []string,
	sem *semaphore.Weighted,
	proxy config.Proxy,
	excludes config.Excludes,
) *Subscription {
	return &Subscription{
		name:          name,
		urlGenerator:  urlGen,
		playlists:     playlists,
		epgs:          epgs,
		semaphore:     sem,
		proxyConfig:   proxy,
		excludeConfig: excludes,
	}
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

func (s *Subscription) StreamCommand(streamUrl string) video.StreamerConfig {
	templVars := make(map[string]string)
	if s.proxyConfig.Stream.TemplateVars != nil {
		for k, v := range s.proxyConfig.Stream.TemplateVars {
			templVars[k] = v
		}
	}
	templVars["url"] = streamUrl

	return video.StreamerConfig{
		Command:      s.proxyConfig.Stream.Command,
		EnvVars:      s.proxyConfig.Stream.EnvVars,
		TemplateVars: templVars,
	}
}

func (s *Subscription) LimitCommand() video.StreamerConfig {
	return video.StreamerConfig{
		Command:      s.proxyConfig.Error.RateLimitExceeded.Command,
		EnvVars:      s.proxyConfig.Error.RateLimitExceeded.EnvVars,
		TemplateVars: s.proxyConfig.Error.RateLimitExceeded.TemplateVars,
	}
}

func (s *Subscription) UpstreamErrorCommand() video.StreamerConfig {
	return video.StreamerConfig{
		Command:      s.proxyConfig.Error.UpstreamError.Command,
		EnvVars:      s.proxyConfig.Error.UpstreamError.EnvVars,
		TemplateVars: s.proxyConfig.Error.UpstreamError.TemplateVars,
	}
}

func (s *Subscription) ExpiredCommand() video.StreamerConfig {
	return video.StreamerConfig{
		Command:      s.proxyConfig.Error.LinkExpired.Command,
		EnvVars:      s.proxyConfig.Error.LinkExpired.EnvVars,
		TemplateVars: s.proxyConfig.Error.LinkExpired.TemplateVars,
	}
}
