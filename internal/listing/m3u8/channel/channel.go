package channel

import (
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
	"net/url"
)

type Subscription interface {
	GetName() string
	GetPlaylists() []string
	GetEPGs() []string
	IsProxied() bool
}

type URLGenerator interface {
	CreateURL(data urlgen.Data) (*url.URL, error)
	Decrypt(s string) (*urlgen.Data, error)
}

type RulesEngine interface {
	ProcessTrack(track *m3u8.Track) bool
	ProcessRegistry(global *Registry, subs *Registry)
}

type Channel struct {
	track        *m3u8.Track
	subscription Subscription
	urlGenerator URLGenerator
	rulesEngine  RulesEngine
}

func New(track *m3u8.Track, subscription Subscription) *Channel {
	return &Channel{
		track:        track,
		subscription: subscription,
	}
}

func NewWithDependencies(track *m3u8.Track, subscription Subscription, urlGenerator URLGenerator, rulesEngine RulesEngine) *Channel {
	return &Channel{
		track:        track,
		subscription: subscription,
		urlGenerator: urlGenerator,
		rulesEngine:  rulesEngine,
	}
}

func (c *Channel) Track() *m3u8.Track {
	return c.track
}

func (c *Channel) Subscription() Subscription {
	return c.subscription
}

func (c *Channel) URLGenerator() URLGenerator {
	return c.urlGenerator
}

func (c *Channel) RulesEngine() RulesEngine {
	return c.rulesEngine
}

func (c *Channel) Name() string {
	return c.track.Name
}

func (c *Channel) ID() string {
	if id, exists := c.track.Attrs[m3u8.AttrTvgID]; exists {
		return id
	}
	return ""
}
