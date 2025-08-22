package m3u8

import (
	"io"
	"iptv-gateway/internal/listing"
)

type decoderWrapper struct {
	decoder      listing.Decoder
	subscription listing.Subscription
	reader       io.ReadCloser
	done         bool
	err          error
}

func (d *decoderWrapper) Close() error {
	if d.reader != nil {
		return d.reader.Close()
	}
	return nil
}
