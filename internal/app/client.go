package app

import (
	"fmt"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/urlgen"

	"golang.org/x/sync/semaphore"
)

type Client struct {
	name          string
	semaphore     *semaphore.Weighted
	subscriptions []*Subscription
	presets       []config.Preset
	proxy         config.Proxy
	rules         []rules.ChannelRule
	playlistRules []rules.PlaylistRule
	epgLink       string
	secret        string
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

func (c *Client) BuildSubscription(
	name string, conf config.Subscription, urlGen urlgen.Generator,
	globalChannelRules []rules.ChannelRule, globalPlaylistUser []rules.PlaylistRule,
	serverProxy config.Proxy,
	sem *semaphore.Weighted) error {

	proxy := mergeProxies(serverProxy, conf.Proxy)
	mergedChannelRules := mergeArrays(globalChannelRules, conf.ChannelRules)
	mergedPlaylistRules := mergeArrays(globalPlaylistUser, conf.PlaylistRules)

	for _, preset := range c.presets {
		proxy = mergeProxies(proxy, preset.Proxy)
		mergedChannelRules = mergeArrays(mergedChannelRules, preset.ChannelRules)
		mergedPlaylistRules = mergeArrays(globalPlaylistUser, preset.PlaylistRules)
	}

	proxy = mergeProxies(proxy, c.proxy)
	mergedChannelRules = mergeArrays(mergedChannelRules, c.rules)
	mergedPlaylistRules = mergeArrays(globalPlaylistUser, c.playlistRules)

	subscription, err := NewSubscription(
		name,
		urlGen,
		conf.Playlist,
		conf.EPG,
		proxy,
		mergedChannelRules,
		mergedPlaylistRules,
		sem,
	)

	if err != nil {
		return err
	}

	c.subscriptions = append(c.subscriptions, subscription)
	return nil
}

func (c *Client) GetEpgLink() string {
	return c.epgLink
}

func (c *Client) GetName() string {
	return c.name
}

func (c *Client) GetSubscriptions() []*Subscription {
	return c.subscriptions
}

func (c *Client) GetSemaphore() *semaphore.Weighted {
	return c.semaphore
}
