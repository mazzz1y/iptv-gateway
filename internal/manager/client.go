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
	proxy         config.Proxy
	excludes      config.Excludes
	epgLink       string
	secret        string
}

func NewClient(clientConfig config.Client, publicUrl string) (*Client, error) {
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
		proxy:     clientConfig.Proxy,
		secret:    clientConfig.Secret,
		epgLink:   fmt.Sprintf("%s/%s/epg.xml.gz", publicUrl, clientConfig.Secret),
	}, nil
}

func (c *Client) AddSubscription(conf config.Subscription, urlGen URLGenerator,
	serverExcludes config.Excludes, serverProxy config.Proxy, sem *semaphore.Weighted) {

	sub := NewSubscription(
		conf.Name,
		urlGen,
		conf.Playlist,
		conf.EPG,
		sem,
		mergeProxies(serverProxy, conf.Proxy, c.proxy),
		mergeExcludes(serverExcludes, conf.Excludes, c.excludes),
	)

	c.subscriptions = append(c.subscriptions, sub)
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
