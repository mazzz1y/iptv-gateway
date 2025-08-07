package manager

import (
	"fmt"
	"golang.org/x/sync/semaphore"
	"iptv-gateway/internal/config"
)

type Client struct {
	name          string
	semaphore     *semaphore.Weighted
	subscriptions []*Subscription
	presets       []config.Preset
	proxy         config.Proxy
	excludes      config.Excludes
	epgLink       string
	secret        string
}

func NewClient(clientConfig config.Client, presets []config.Preset, publicUrl string) (*Client, error) {
	if clientConfig.Secret == "" {
		return nil, fmt.Errorf("client secret cannot be empty")
	}

	var sem *semaphore.Weighted
	if clientConfig.Proxy.ConcurrentStreams > 0 {
		sem = semaphore.NewWeighted(clientConfig.Proxy.ConcurrentStreams)
	}

	return &Client{
		name:      clientConfig.Name,
		semaphore: sem,
		presets:   presets,
		proxy:     clientConfig.Proxy,
		secret:    clientConfig.Secret,
		epgLink:   fmt.Sprintf("%s/%s/epg.xml.gz", publicUrl, clientConfig.Secret),
	}, nil
}

func (c *Client) AddSubscription(
	conf config.Subscription, urlGen URLGenerator,
	serverExcludes config.Excludes, serverProxy config.Proxy,
	sem *semaphore.Weighted) error {

	proxy := mergeProxies(serverProxy, conf.Proxy)
	exclude := mergeExcludes(serverExcludes, conf.Excludes)

	for _, preset := range c.presets {
		proxy = mergeProxies(proxy, preset.Proxy)
		exclude = mergeExcludes(exclude, preset.Excludes)
	}

	proxy = mergeProxies(proxy, c.proxy)
	exclude = mergeExcludes(exclude, c.excludes)

	sub, err := NewSubscription(
		conf.Name,
		urlGen,
		conf.Playlist,
		conf.EPG,
		proxy,
		exclude,
		sem,
	)

	if err != nil {
		return err
	}

	c.subscriptions = append(c.subscriptions, sub)
	return nil
}

func (c *Client) GetEpgLink() string {
	return c.epgLink
}

func (c *Client) IsCorrectSecret(secret string) bool {
	return c.secret == secret
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
