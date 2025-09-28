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
	proxy             proxy.Proxy
	rulesProcessor    *m3u8Rules.Processor
	epgLink           string
	urlGen            *urlgen.Generator
}

type Provider interface {
	Name() string
	Type() string
	URLGenerator() *urlgen.Generator
	ExpiredLinkStreamer() *shell.Streamer
}

func NewClient(clientCfg config.Client, urlGen *urlgen.Generator, channelRules []*rules.ChannelRule, playlistRules []*rules.PlaylistRule, publicURL string) (*Client, error) {
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
		proxy:          clientCfg.Proxy,
		rulesProcessor: m3u8Rules.NewProcessor(clientCfg.Name, channelRules, playlistRules),
		epgLink:        fmt.Sprintf("%s/%s/epg.xml.gz", publicURL, clientCfg.Secret),
		urlGen:         urlGen,
	}, nil
}

func (c *Client) BuildPlaylistProvider(
	playlistConf config.Playlist, serverProxy proxy.Proxy, sem *semaphore.Weighted) error {

	pr, err := NewPlaylistProvider(
		playlistConf.Name,
		c.urlGen,
		playlistConf.Sources,
		mergeProxies(serverProxy, playlistConf.Proxy, c.proxy),
		nil,
		sem,
	)
	if err != nil {
		return err
	}

	c.playlistProviders = append(c.playlistProviders, pr)
	return nil
}

func (c *Client) BuildEPGProvider(
	epgConf config.EPG, serverProxy proxy.Proxy) error {

	subscription, err := NewEPGProvider(
		epgConf.Name,
		c.urlGen,
		epgConf.Sources,
		mergeProxies(serverProxy, epgConf.Proxy, c.proxy),
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

func (c *Client) GetProvider(prType urlgen.ProviderType, prName string) Provider {
	switch prType {
	case urlgen.ProviderTypePlaylist:
		for _, ps := range c.playlistProviders {
			if prName == ps.Name() {
				return ps
			}
		}
	case urlgen.ProviderTypeEPG:
		for _, ps := range c.epgProviders {
			if prName == ps.Name() {
				return ps
			}
		}
	}

	return nil
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

func (c *Client) URLGenerator() *urlgen.Generator {
	return c.urlGen
}

func (c *Client) RulesProcessor() *m3u8Rules.Processor {
	return c.rulesProcessor
}
