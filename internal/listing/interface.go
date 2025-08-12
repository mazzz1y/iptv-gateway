package listing

import (
	"context"
	"iptv-gateway/internal/cache"
)

type HTTPClient interface {
	NewReader(ctx context.Context, url string) (*cache.Reader, error)
}

type Decoder interface {
	Decode() (any, error)
	Close() error
}
