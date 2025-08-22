package listing

import (
	"iptv-gateway/internal/urlgen"
	"net/http"
	"net/url"
	"time"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type Decoder interface {
	Decode() (any, error)
}

type URLGenerator interface {
	CreateURL(data urlgen.Data, ttl time.Duration) (*url.URL, error)
}
