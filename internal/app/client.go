package app

import (
	"fmt"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/listing"
	m3u8Rules "iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/shell"
	"iptv-gateway/internal/urlgen"

	"golang.org/x/sync/semaphore"
)

type Client struct {
	name              string
	secret            string
	semaphore         *semaphore.Weighted
	playlistProviders []*Playlist
	epgProviders      []*EPG
	presets           []config.Preset
	proxy             proxy.Proxy
	rules             []*rules.Rule
	rulesProcessor    *m3u8Rules.Processor
	epgLink           string
}

type Provider interface {
	Name() string
	Type() string
	URLGenerator() *urlgen.Generator
	ExpiredLinkStreamer() *shell.Streamer
}

func NewClient(clientCfg config.Client, presets []config.Preset, publicURL string) (*Client, error) {
	if clientCfg.Secret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}

	var sem *semaphore.Weighted
	if clientCfg.Proxy.ConcurrentStreams > 0 {
		sem = semaphore.NewWeighted(clientCfg.Proxy.ConcurrentStreams)
	}

	return &Client{
		name:           clientCfg.Name,
		secret:         clientCfg.Secret,
		semaphore:      sem,
		presets:        presets,
		proxy:          clientCfg.Proxy,
		rules:          clientCfg.Rules,
		rulesProcessor: m3u8Rules.NewProcessor(),
		epgLink:        fmt.Sprintf("%s/%s/epg.xml.gz", publicURL, clientCfg.Secret),
	}, nil
}

func (c *Client) BuildPlaylistProvider(
	playlistConf config.Playlist, urlGen urlgen.Generator,
	globalRules []*rules.Rule, serverProxy proxy.Proxy, sem *semaphore.Weighted) error {

	var mergedRules []*rules.Rule
	mergedRules = append(mergedRules, playlistConf.Rules...)
	mergedRules = append(mergedRules, presetRules(c.presets)...)
	mergedRules = append(mergedRules, c.rules...)
	mergedRules = append(mergedRules, globalRules...)

	pr, err := NewPlaylistProvider(
		playlistConf.Name,
		urlGen,
		playlistConf.Sources,
		mergeProxies(serverProxy, playlistConf.Proxy, presetProxy(c.presets), c.proxy),
		mergedRules,
		sem,
	)
	if err != nil {
		return err
	}

	c.rulesProcessor.AddPlaylist(pr)
	c.playlistProviders = append(c.playlistProviders, pr)
	return nil
}

func (c *Client) BuildEPGProvider(
	epgConf config.EPG, urlGen urlgen.Generator, serverProxy proxy.Proxy) error {

	subscription, err := NewEPGProvider(
		epgConf.Name,
		urlGen,
		epgConf.Sources,
		mergeProxies(serverProxy, epgConf.Proxy, presetProxy(c.presets), c.proxy),
	)
	if err != nil {
		return err
	}

	c.epgProviders = append(c.epgProviders, subscription)
	return nil
}

func (c *Client) PlaylistProviders() []listing.Playlist {
	result := make([]listing.Playlist, 0, len(c.playlistProviders))
	for _, ps := range c.playlistProviders {
		result = append(result, ps)
	}
	return result
}

func (c *Client) EPGProviders() []listing.EPG {
	result := make([]listing.EPG, 0, len(c.epgProviders))
	for _, es := range c.epgProviders {
		result = append(result, es)
	}
	return result
}

func (c *Client) Providers() []Provider {
	providers := make([]Provider, 0, len(c.playlistProviders)+len(c.epgProviders))
	for _, ps := range c.playlistProviders {
		providers = append(providers, ps)
	}
	for _, es := range c.epgProviders {
		providers = append(providers, es)
	}
	return providers
}

func (c *Client) Semaphore() *semaphore.Weighted {
	return c.semaphore
}

func (c *Client) EPGLink() string {
	return c.epgLink
}

func (c *Client) Name() string {
	return c.name
}

func presetRules(presets []config.Preset) []*rules.Rule {
	var rs []*rules.Rule
	for _, preset := range presets {
		rs = append(rs, preset.Rules...)
	}
	return rs
}

func presetProxy(presets []config.Preset) proxy.Proxy {
	var result proxy.Proxy
	for _, preset := range presets {
		result = mergeProxies(result, preset.Proxy)
	}
	return result
}
