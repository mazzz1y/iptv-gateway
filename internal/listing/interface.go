package listing

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/urlgen"
	"net/http"
	"net/url"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Decoder interface {
	Decode() (any, error)
}

type URLGenerator interface {
	CreateURL(data urlgen.Data, ttl time.Duration) (*url.URL, error)
}

type PlaylistSubscription interface {
	Playlists() []string
	URLGenerator() *urlgen.Generator
	ChannelRules() []rules.ChannelRule
	PlaylistRules() []rules.PlaylistRule
	IsProxied() bool
}

type EPGSubscription interface {
	EPGs() []string
	URLGenerator() *urlgen.Generator
	IsProxied() bool
}
