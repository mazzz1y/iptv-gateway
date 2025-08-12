package listing

import (
	"context"
	"iptv-gateway/internal/cache"
)

type HTTPClient interface {
	NewReader(ctx context.Context, url string) (*cache.Reader, error)
	Close()
}

type Decoder interface {
	Decode() (any, error)
}
