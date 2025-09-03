package xmltv

import (
	"io"
	"iptv-gateway/internal/listing"
)

type decoderWrapper struct {
	decoder         listing.Decoder
	subscription    listing.EPGSubscription
	reader          io.ReadCloser
	channelsDone    bool
	done            bool
	err             error
	bufferedItem    any
	hasBufferedItem bool
}

func (d *decoderWrapper) nextItem() (any, error) {
	if d.hasBufferedItem {
		item := d.bufferedItem
		d.bufferedItem = nil
		d.hasBufferedItem = false
		return item, nil
	}
	return d.decoder.Decode()
}

func (d *decoderWrapper) Close() error {
	if d.reader != nil {
		return d.reader.Close()
	}
	return nil
}
