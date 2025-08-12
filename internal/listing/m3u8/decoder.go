package m3u8

import (
	"iptv-gateway/internal/manager"
)

type Decoder interface {
	Decode() (any, error)
	Close() error
}

type decoderWrapper struct {
	decoder      Decoder
	subscription *manager.Subscription
	done         bool
	err          error
}

func (dw *decoderWrapper) close() error {
	if dw.decoder != nil {
		return dw.decoder.Close()
	}
	return nil
}
