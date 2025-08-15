package m3u8

import (
	"context"
	"io"
	"iptv-gateway/internal/client"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/listing/m3u8/channel"
)

type Streamer struct {
	subscriptions []*client.Subscription
	httpClient    listing.HTTPClient
	epgLink       string
}

func NewStreamer(subs []*client.Subscription, epgLink string, httpClient listing.HTTPClient) *Streamer {
	return &Streamer{
		subscriptions: subs,
		httpClient:    httpClient,
		epgLink:       epgLink,
	}
}

func (s *Streamer) WriteTo(ctx context.Context, w io.Writer) (int64, error) {
	registry, err := s.getChannels(ctx)
	if err != nil {
		return 0, err
	}

	writer := NewWriter(s.epgLink)
	return writer.WriteRegistry(registry, w)
}

func (s *Streamer) GetAllChannels(ctx context.Context) (map[string]bool, error) {
	registry, err := s.getChannels(ctx)
	if err != nil {
		return nil, err
	}

	channels := make(map[string]bool)
	for _, ch := range registry.All() {
		if id := ch.ID(); id != "" {
			channels[id] = true
		}
		channels[ch.Name()] = true
	}

	return channels, nil
}

func (s *Streamer) getChannels(ctx context.Context) (*channel.Registry, error) {
	fetcher := NewPlaylistFetcher(s.subscriptions, s.httpClient)
	registry, err := fetcher.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	processor := NewProcessor()
	if err := processor.Process(registry); err != nil {
		return nil, err
	}

	return registry, nil
}
