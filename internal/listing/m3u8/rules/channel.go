package rules

import (
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
	"net/url"
	"time"
)

type Subscription interface {
	IsProxied() bool
}

type URLGenerator interface {
	CreateURL(data urlgen.Data, ttl time.Duration) (*url.URL, error)
}

type Channel struct {
	track        *m3u8.Track
	subscription Subscription
}

func NewChannel(track *m3u8.Track, subscription Subscription) *Channel {
	return &Channel{
		track:        track,
		subscription: subscription,
	}
}

func (c *Channel) Track() *m3u8.Track {
	return c.track
}

func (c *Channel) Subscription() Subscription {
	return c.subscription
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
