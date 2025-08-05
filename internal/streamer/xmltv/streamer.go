package xmltv

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"iptv-gateway/internal/cache"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/ioutils"
	"iptv-gateway/internal/manager"
	"iptv-gateway/internal/streamer/common"
	"iptv-gateway/internal/url_generator"
	"iptv-gateway/internal/xmltv"
	"syscall"
	"time"
)

type xmltvDecoderFactory struct{}

func (f *xmltvDecoderFactory) NewDecoder(reader *cache.Reader) common.Decoder {
	return xmltv.NewDecoder(reader)
}

type Streamer struct {
	*common.BaseStreamer
	channels            map[string]bool
	addedChannels       map[string]bool
	addedProgrammes     map[string]bool
	currentURLGenerator *url_generator.Generator
}

func NewStreamer(subscriptions []*manager.Subscription, httpClient common.HTTPClient, channels map[string]bool) *Streamer {
	decoderFactory := &xmltvDecoderFactory{}
	baseStreamer := common.NewBaseStreamer(subscriptions, httpClient, decoderFactory)

	return &Streamer{
		BaseStreamer:    baseStreamer,
		channels:        channels,
		addedChannels:   make(map[string]bool, 10000),
		addedProgrammes: make(map[string]bool, 100000),
	}
}

func (s *Streamer) WriteToGzip(ctx context.Context, w io.Writer) (int64, error) {
	gzWriter, _ := gzip.NewWriterLevel(w, constant.GzipLevel)
	defer gzWriter.Close()

	return s.WriteTo(ctx, gzWriter)
}

func (s *Streamer) WriteTo(ctx context.Context, w io.Writer) (int64, error) {
	if len(s.Subscriptions) == 0 {
		return 0, fmt.Errorf("no EPG sources found")
	}

	bytesCounter := ioutils.NewCountWriter(w)
	encoder := xmltv.NewEncoder(bytesCounter)
	defer encoder.Close()

	s.PendingSubscriptions = s.Subscriptions
	s.CurrentDecoder = nil
	s.Close()

	getEPGs := func(sub *manager.Subscription) []string {
		return sub.GetEPGs()
	}

	for {
		item, err := s.NextItem(ctx, getEPGs)
		if err == io.EOF {
			break
		}
		if err != nil {
			return bytesCounter.Count(), err
		}

		if s.CurrentSubscription != nil && s.CurrentSubscription.IsProxied() {
			s.currentURLGenerator = s.CurrentSubscription.GetURLGenerator().(*url_generator.Generator)
		} else {
			s.currentURLGenerator = nil
		}

		switch v := item.(type) {
		case xmltv.Channel:
			if !s.isChannelAllowed(v) {
				continue
			}
			v.Icons = s.processIcons(v.Icons)

			if err := encoder.Encode(v); err != nil {
				return bytesCounter.Count(), err
			}

		case xmltv.Programme:
			if !s.isProgrammeAllowed(v) {
				continue
			}
			v.Icons = s.processIcons(v.Icons)

			err := encoder.Encode(v)
			if errors.Is(err, syscall.EPIPE) {
				return bytesCounter.Count(), nil
			}
			if err != nil {
				return bytesCounter.Count(), err
			}
		}
	}

	count := bytesCounter.Count()
	if count == 0 {
		return count, fmt.Errorf("no data in subscriptions")
	}

	return count, nil
}

func (s *Streamer) isChannelAllowed(channel xmltv.Channel) bool {
	if s.addedChannels[channel.ID] {
		return false
	}

	if s.channels[channel.ID] {
		s.addedChannels[channel.ID] = true
		return true
	}

	for _, displayName := range channel.DisplayNames {
		if s.channels[displayName.Value] {
			s.addedChannels[channel.ID] = true
			return true
		}
	}

	return false
}

func (s *Streamer) isProgrammeAllowed(programme xmltv.Programme) bool {
	if !s.addedChannels[programme.Channel] {
		return false
	}

	key := programme.Channel
	if programme.Start != nil {
		key += programme.Start.Time.Format(time.RFC3339)
	}
	if programme.ID != "" {
		key += programme.ID
	}

	if s.addedProgrammes[key] {
		return false
	}

	s.addedProgrammes[key] = true
	return true
}

func (s *Streamer) processIcons(icons []xmltv.Icon) []xmltv.Icon {
	if s.currentURLGenerator == nil || len(icons) == 0 {
		return icons
	}

	needsUpdate := false
	for _, icon := range icons {
		if icon.Source != "" {
			needsUpdate = true
			break
		}
	}

	if !needsUpdate {
		return icons
	}

	result := make([]xmltv.Icon, len(icons))
	copy(result, icons)

	urlData := url_generator.Data{
		RequestType: url_generator.File,
	}

	for i := range result {
		if result[i].Source == "" {
			continue
		}

		urlData.URL = result[i].Source
		link, err := s.currentURLGenerator.CreateURL(urlData)
		if err != nil {
			continue
		}
		result[i].Source = link.String()
	}

	return result
}
