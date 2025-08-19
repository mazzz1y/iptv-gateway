package m3u8

import (
	"errors"
	"io"
	"iptv-gateway/internal/ioutil"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
	"syscall"
)

type Writer struct {
	epgLink string
}

func NewWriter(epgLink string) *Writer {
	return &Writer{epgLink: epgLink}
}

func (w *Writer) WriteChannels(channels []*rules.Channel, writer io.Writer) (int64, error) {
	bytesCounter := ioutil.NewCountWriter(writer)
	encoder := m3u8.NewEncoder(bytesCounter, map[string]string{"x-tvg-url": w.epgLink})
	defer encoder.Close()

	for _, ch := range channels {
		err := encoder.Encode(ch.Track())
		if errors.Is(err, syscall.EPIPE) {
			break
		}
		if err != nil {
			return bytesCounter.Count(), err
		}
	}

	return bytesCounter.Count(), nil
}
