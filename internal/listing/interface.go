package listing

import (
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/shell"
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

type Subscription interface {
	GetPlaylists() []string
	GetEPGs() []string
	GetURLGenerator() *urlgen.Generator
	GetRules() []config.RuleAction
	IsProxied() bool
	GetName() string
	ExpiredCommandStreamer() *shell.Streamer
}
