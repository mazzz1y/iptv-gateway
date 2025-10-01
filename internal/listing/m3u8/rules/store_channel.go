package rules

import (
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
	"net/url"
	"time"
)

type URLGenerator interface {
	CreateURL(data urlgen.Data, ttl time.Duration) (*url.URL, error)
}

type Channel struct {
	track    *m3u8.Track
	playlist listing.Playlist
	hidden   bool
	removed  bool
}

func NewChannel(track *m3u8.Track, playlist listing.Playlist) *Channel {
	return &Channel{
		track:    track,
		playlist: playlist,
	}
}

func (c *Channel) Track() *m3u8.Track {
	return c.track
}

func (c *Channel) Playlist() listing.Playlist {
	return c.playlist
}

func (c *Channel) Name() string {
	return c.track.Name
}

func (c *Channel) ID() string {
	if id, exists := c.GetAttr(m3u8.AttrTvgID); exists {
		return id
	}
	return ""
}

func (c *Channel) URI() *url.URL {
	return c.track.URI
}

func (c *Channel) SetName(name string) {
	c.track.Name = name
}

func (c *Channel) SetURI(uri *url.URL) {
	c.track.URI = uri
}

func (c *Channel) IsHidden() bool {
	return c.hidden
}

func (c *Channel) IsRemoved() bool {
	return c.removed
}

func (c *Channel) MarkHidden() {
	c.hidden = true
}

func (c *Channel) MarkRemoved() {
	c.removed = true
}

func (c *Channel) Attrs() map[string]string {
	return c.track.Attrs
}

func (c *Channel) GetAttr(key string) (string, bool) {
	if c.track.Attrs == nil {
		return "", false
	}
	value, exists := c.track.Attrs[key]
	return value, exists
}

func (c *Channel) SetAttr(key, value string) {
	if c.track.Attrs == nil {
		c.track.Attrs = make(map[string]string)
	}
	c.track.Attrs[key] = value
}

func (c *Channel) DeleteAttr(key string) {
	if c.track.Attrs != nil {
		delete(c.track.Attrs, key)
	}
}

func (c *Channel) Tags() map[string]string {
	return c.track.Tags
}

func (c *Channel) GetTag(key string) (string, bool) {
	if c.track.Tags == nil {
		return "", false
	}
	value, exists := c.track.Tags[key]
	return value, exists
}

func (c *Channel) SetTag(key, value string) {
	if c.track.Tags == nil {
		c.track.Tags = make(map[string]string)
	}
	c.track.Tags[key] = value
}

func (c *Channel) DeleteTag(key string) {
	if c.track.Tags != nil {
		delete(c.track.Tags, key)
	}
}
