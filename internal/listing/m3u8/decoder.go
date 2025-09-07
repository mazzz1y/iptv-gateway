package m3u8

import (
	"context"
	"io"

	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/parser/m3u8"
)

type decoderWrapper struct {
	*listing.BaseDecoder
	subscription listing.Playlist
}

func newDecoderWrapper(subscription listing.Playlist, httpClient listing.HTTPClient, url string) *decoderWrapper {
	initFunc := func(ctx context.Context, url string) (listing.Decoder, io.ReadCloser, error) {
		reader, err := listing.CreateReader(ctx, httpClient, url)
		if err != nil {
			return nil, nil, err
		}
		decoder := m3u8.NewDecoder(reader)
		return decoder, reader, nil
	}

	return &decoderWrapper{
		BaseDecoder:  listing.NewLazyBaseDecoder(url, initFunc),
		subscription: subscription,
	}
}
