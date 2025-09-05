package app

import (
	"fmt"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/shell"
	"iptv-gateway/internal/urlgen"

	"golang.org/x/sync/semaphore"
)

type Client struct {
	name                  string
	semaphore             *semaphore.Weighted
	playlistSubscriptions []*PlaylistSubscription
	epgSubscriptions      []*EPGSubscription
	presets               []config.Preset
	proxy                 config.Proxy
	rules                 []rules.ChannelRule
	playlistRules         []rules.PlaylistRule
	epgLink               string
	secret                string
}

type ProviderWithURLGen struct {
	Provider        Provider
	URLGen          *urlgen.Generator
	ExpiredStreamer *shell.Streamer
}

type Provider interface {
	Name() string
	Type() string
}

func NewClient(name string, clientCfg config.Client, presets []config.Preset, publicUrl string) (*Client, error) {
	if clientCfg.Secret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}

	var sem *semaphore.Weighted
	if clientCfg.Proxy.ConcurrentStreams > 0 {
		sem = semaphore.NewWeighted(clientCfg.Proxy.ConcurrentStreams)
	}

	return &Client{
		name:          name,
		semaphore:     sem,
		presets:       presets,
		proxy:         clientCfg.Proxy,
		secret:        clientCfg.Secret,
		rules:         clientCfg.ChannelRules,
		playlistRules: clientCfg.PlaylistRules,
		epgLink:       fmt.Sprintf("%s/%s/epg.xml.gz", publicUrl, clientCfg.Secret),
	}, nil
}

func (c *Client) BuildPlaylistSubscription(
	playlistConf config.Playlist, urlGen urlgen.Generator,
	globalChannelRules []rules.ChannelRule, globalPlaylistUser []rules.PlaylistRule,
	serverProxy config.Proxy,
	sem *semaphore.Weighted,
	namedConditions []rules.NamedCondition) error {

	playlistProxy := mergeProxies(serverProxy, playlistConf.Proxy)
	mergedChannelRules := mergeArrays(globalChannelRules, playlistConf.ChannelRules)
	mergedPlaylistRules := mergeArrays(globalPlaylistUser, playlistConf.PlaylistRules)

	for _, preset := range c.presets {
		playlistProxy = mergeProxies(playlistProxy, preset.Proxy)
		mergedChannelRules = mergeArrays(mergedChannelRules, preset.ChannelRules)
		mergedPlaylistRules = mergeArrays(mergedPlaylistRules, preset.PlaylistRules)
	}

	playlistProxy = mergeProxies(playlistProxy, c.proxy)
	mergedChannelRules = mergeArrays(mergedChannelRules, c.rules)
	mergedPlaylistRules = mergeArrays(mergedPlaylistRules, c.playlistRules)

	subscription, err := NewPlaylistSubscription(
		playlistConf.Name,
		urlGen,
		playlistConf.Sources,
		playlistProxy,
		mergedChannelRules,
		mergedPlaylistRules,
		sem,
		namedConditions,
	)

	if err != nil {
		return err
	}

	c.playlistSubscriptions = append(c.playlistSubscriptions, subscription)
	return nil
}

func (c *Client) BuildEPGSubscription(
	epgConf config.EPG, urlGen urlgen.Generator,
	serverProxy config.Proxy) error {

	epgProxy := mergeProxies(serverProxy, epgConf.Proxy)

	for _, preset := range c.presets {
		epgProxy = mergeProxies(epgProxy, preset.Proxy)
	}

	epgProxy = mergeProxies(epgProxy, c.proxy)

	subscription, err := NewEPGSubscription(
		epgConf.Name,
		urlGen,
		epgConf.Sources,
		epgProxy,
	)

	if err != nil {
		return err
	}

	c.epgSubscriptions = append(c.epgSubscriptions, subscription)
	return nil
}

func (c *Client) EpgLink() string {
	return c.epgLink
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) PlaylistSubscriptions() []listing.Playlist {
	result := make([]listing.Playlist, len(c.playlistSubscriptions))
	for i, ps := range c.playlistSubscriptions {
		result[i] = ps
	}
	return result
}

func (c *Client) EPGSubscriptions() []listing.EPG {
	result := make([]listing.EPG, len(c.epgSubscriptions))
	for i, es := range c.epgSubscriptions {
		result[i] = es
	}
	return result
}

func (c *Client) Semaphore() *semaphore.Weighted {
	return c.semaphore
}

func (c *Client) URLProviders() <-chan ProviderWithURLGen {
	totalProviders := len(c.playlistSubscriptions) + len(c.epgSubscriptions)
	ch := make(chan ProviderWithURLGen, totalProviders)

	go func() {
		defer close(ch)

		for _, ps := range c.playlistSubscriptions {
			ch <- ProviderWithURLGen{
				Provider:        ps,
				URLGen:          ps.urlGenerator,
				ExpiredStreamer: ps.expiredLinkStreamer,
			}
		}

		for _, es := range c.epgSubscriptions {
			ch <- ProviderWithURLGen{
				Provider:        es,
				URLGen:          es.urlGenerator,
				ExpiredStreamer: nil,
			}
		}
	}()

	return ch
}
