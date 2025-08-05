package m3u8

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iptv-gateway/internal/cache"
	"iptv-gateway/internal/ioutils"
	"iptv-gateway/internal/m3u8"
	"iptv-gateway/internal/manager"
	"iptv-gateway/internal/streamer/common"
	"iptv-gateway/internal/url_generator"
	"net/url"
	"strings"
	"syscall"
)

type m3u8DecoderFactory struct{}

func (f *m3u8DecoderFactory) NewDecoder(reader *cache.Reader) common.Decoder {
	return m3u8.NewDecoder(reader)
}

type Streamer struct {
	*common.BaseStreamer
	epgLink         string
	addedTrackIDs   map[string]bool
	addedTrackNames map[string]bool
}

func NewStreamer(subs []*manager.Subscription, epgLink string, httpClient common.HTTPClient) *Streamer {
	decoderFactory := &m3u8DecoderFactory{}
	baseStreamer := common.NewBaseStreamer(subs, httpClient, decoderFactory)

	return &Streamer{
		BaseStreamer:    baseStreamer,
		epgLink:         epgLink,
		addedTrackIDs:   make(map[string]bool),
		addedTrackNames: make(map[string]bool),
	}
}

func (s *Streamer) WriteTo(ctx context.Context, w io.Writer) (int64, error) {
	if len(s.Subscriptions) == 0 {
		return 0, fmt.Errorf("no subscriptions found")
	}

	bytesCounter := ioutils.NewCountWriter(w)
	encoder := m3u8.NewEncoder(bytesCounter, map[string]string{"x-tvg-url": s.epgLink})
	defer encoder.Close()

	s.PendingSubscriptions = s.Subscriptions
	s.CurrentDecoder = nil
	s.Close()

	getPlaylists := func(sub *manager.Subscription) []string {
		return sub.GetPlaylists()
	}

	for {
		item, err := s.NextItem(ctx, getPlaylists)

		if err == io.EOF {
			break
		}
		if err != nil {
			return bytesCounter.Count(), err
		}

		if track, ok := item.(*m3u8.Track); ok {
			if s.isDuplicate(track) {
				continue
			}

			if s.isExcluded(track) {
				continue
			}

			if s.CurrentSubscription.IsProxied() {
				if err := s.processProxyLinks(track); err != nil {
					return bytesCounter.Count(), err
				}
			}
			err := encoder.Encode(track)
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

func (s *Streamer) GetAllChannels(ctx context.Context) (map[string]bool, error) {
	if len(s.Subscriptions) == 0 {
		return nil, fmt.Errorf("no subscriptions found")
	}

	channels := make(map[string]bool)
	s.PendingSubscriptions = s.Subscriptions
	s.CurrentDecoder = nil
	s.Close()

	getPlaylists := func(sub *manager.Subscription) []string {
		return sub.GetPlaylists()
	}

	for {
		item, err := s.NextItem(ctx, getPlaylists)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if track, ok := item.(*m3u8.Track); ok {
			if id, hasID := track.Attrs[m3u8.AttrTvgID]; hasID && id != "" {
				channels[id] = true
			}
			channels[track.Name] = true
		}
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels found in subscriptions")
	}

	return channels, nil
}

func (s *Streamer) isExcluded(track *m3u8.Track) bool {
	filters := s.CurrentSubscription.GetExcludes()

	if len(filters.Tags) == 0 && len(filters.Attrs) == 0 && len(filters.ChannelName) == 0 {
		return false
	}

	for _, pattern := range filters.ChannelName {
		if pattern.MatchString(track.Name) {
			return true
		}
	}

	for attrKey, patterns := range filters.Attrs {
		attrValue, exists := track.Attrs[attrKey]
		if exists {
			for _, pattern := range patterns {
				if pattern.MatchString(attrValue) {
					return true
				}
			}
		}
	}

	for tagKey, patterns := range filters.Tags {
		tagValue, exists := track.Tags[tagKey]
		if exists {
			for _, pattern := range patterns {
				if pattern.MatchString(tagValue) {
					return true
				}
			}
		}
	}

	return false
}

func (s *Streamer) isDuplicate(track *m3u8.Track) bool {
	id, hasID := track.Attrs[m3u8.AttrTvgID]
	trackName := strings.ToLower(track.Name)

	if hasID && id != "" {
		if s.addedTrackIDs[id] {
			return true
		}
		s.addedTrackIDs[id] = true
		return false
	}

	if s.addedTrackNames[trackName] {
		return true
	}
	s.addedTrackNames[trackName] = true
	return false
}

func (s *Streamer) processProxyLinks(track *m3u8.Track) error {
	urlGenerator, ok := s.CurrentSubscription.GetURLGenerator().(*url_generator.Generator)
	if !ok {
		mockGenerator, ok := s.CurrentSubscription.GetURLGenerator().(interface {
			CreateURL(data url_generator.Data) (*url.URL, error)
		})
		if !ok {
			return fmt.Errorf("invalid URL generator type")
		}

		for key, value := range track.Attrs {
			if isURL(value) {
				encURL, err := mockGenerator.CreateURL(url_generator.Data{
					RequestType: url_generator.File,
					URL:         value,
				})
				if err != nil {
					return fmt.Errorf("failed to encode attribute URL: %w", err)
				}
				track.Attrs[key] = encURL.String()
			}
		}

		if track.URI != nil && isURL(track.URI.String()) {
			newURL, err := mockGenerator.CreateURL(url_generator.Data{
				RequestType: url_generator.Stream,
				ChannelID:   track.Name,
				URL:         track.URI.String(),
			})
			if err != nil {
				return fmt.Errorf("failed to encode stream URL: %w", err)
			}
			track.URI = newURL
		}

		return nil
	}

	for key, value := range track.Attrs {
		if isURL(value) {
			encURL, err := urlGenerator.CreateURL(url_generator.Data{
				RequestType: url_generator.File,
				URL:         value,
			})
			if err != nil {
				return fmt.Errorf("failed to encode attribute URL: %w", err)
			}
			track.Attrs[key] = encURL.String()
		}
	}

	if track.URI != nil && isURL(track.URI.String()) {
		newURL, err := urlGenerator.CreateURL(url_generator.Data{
			RequestType: url_generator.Stream,
			ChannelID:   track.Name,
			URL:         track.URI.String(),
		})
		if err != nil {
			return fmt.Errorf("failed to encode stream URL: %w", err)
		}
		track.URI = newURL
	}

	return nil
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Host != ""
}
