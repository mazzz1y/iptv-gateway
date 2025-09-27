package m3u8

import (
	"context"
	"fmt"
	"io"
	"iptv-gateway/internal/listing"
	m3u8Rules "iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
)

type Streamer struct {
	subscriptions  []listing.Playlist
	httpClient     listing.HTTPClient
	epgURL         string
	rulesProcessor *m3u8Rules.Processor
}

func NewStreamer(subs []listing.Playlist, epgLink string, httpClient listing.HTTPClient, rulesProcessor *m3u8Rules.Processor) *Streamer {
	return &Streamer{
		subscriptions:  subs,
		httpClient:     httpClient,
		epgURL:         epgLink,
		rulesProcessor: rulesProcessor,
	}
}

func (s *Streamer) WriteTo(ctx context.Context, w io.Writer) (int64, error) {
	channels, err := s.getChannels(ctx)
	if err != nil {
		return 0, err
	}

	writer := NewWriter(s.epgURL)
	return writer.WriteChannels(channels, w)
}

func (s *Streamer) GetAllChannels(ctx context.Context) (map[string]string, error) {
	channels, err := s.getChannels(ctx)
	if err != nil {
		return nil, err
	}

	channelMap := make(map[string]string)
	for _, ch := range channels {
		if tvgID, exists := ch.GetAttr("tvg-id"); exists {
			channelMap[tvgID] = ch.Name()
		}
	}

	return channelMap, nil
}

func (s *Streamer) getChannels(ctx context.Context) ([]*m3u8Rules.Channel, error) {
	store, err := s.fetchPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	processor := NewProcessor()

	return processor.Process(store, s.rulesProcessor)
}

func (s *Streamer) fetchPlaylists(ctx context.Context) (*m3u8Rules.Store, error) {
	store := m3u8Rules.NewStore()

	var decoders []*decoderWrapper
	for _, sub := range s.subscriptions {
		for _, url := range sub.Playlists() {
			decoders = append(decoders, newDecoderWrapper(sub, s.httpClient, url))
		}
	}

	defer func() {
		for _, decoder := range decoders {
			if decoder != nil {
				decoder.Close()
			}
		}
	}()

	for _, decoder := range decoders {
		err := decoder.StartBuffering(ctx)
		if err != nil {
			return nil, err
		}
	}

	for _, decoder := range decoders {
		if err := s.processTracks(ctx, decoder, store); err != nil {
			return nil, err
		}
	}

	if store.Len() == 0 {
		return nil, fmt.Errorf("no channels found in subscriptions")
	}

	return store, nil
}

func (s *Streamer) processTracks(ctx context.Context, decoder *decoderWrapper, store *m3u8Rules.Store) error {
	decoder.StopBuffer()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			item, err := decoder.NextItem()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			if track, ok := item.(*m3u8.Track); ok {
				ch := m3u8Rules.NewChannel(track, decoder.subscription)
				store.Add(ch)
			}
		}
	}
}
