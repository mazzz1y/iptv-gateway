package m3u8

import (
	"context"
	"fmt"
	"io"

	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
)

type Streamer struct {
	subscriptions []listing.PlaylistSubscription
	httpClient    listing.HTTPClient
	epgURL        string
}

func NewStreamer(subs []listing.PlaylistSubscription, epgLink string, httpClient listing.HTTPClient) *Streamer {
	return &Streamer{
		subscriptions: subs,
		httpClient:    httpClient,
		epgURL:        epgLink,
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
		track := ch.Track()
		channelMap[track.Attrs["tvg-id"]] = track.Name
	}

	return channelMap, nil
}

func (s *Streamer) getChannels(ctx context.Context) ([]*rules.Channel, error) {
	store, err := s.fetchPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	rulesProcessor := s.createRulesProcessor()
	processor := NewProcessor()

	return processor.Process(store, rulesProcessor)
}

func (s *Streamer) fetchPlaylists(ctx context.Context) (*rules.Store, error) {
	store := rules.NewStore()

	decoders, err := s.initDecoders(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, decoder := range decoders {
			decoder.Close()
		}
	}()

	err = listing.Process(
		ctx,
		decoders,
		func(ctx context.Context, decoder *decoderWrapper, output chan<- listing.Item, errChan chan<- error) {
			for {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				default:
				}

				item, err := decoder.decoder.Decode()
				if err == io.EOF {
					return
				}
				if err != nil {
					errChan <- err
					return
				}

				if track, ok := item.(*m3u8.Track); ok {
					ch := rules.NewChannel(track, decoder.subscription)
					store.Add(ch)
				}
			}
		},
		nil,
	)

	if err != nil {
		return nil, err
	}

	if store.Len() == 0 {
		return nil, fmt.Errorf("no channels found in subscriptions")
	}

	return store, nil
}

func (s *Streamer) initDecoders(ctx context.Context) ([]*decoderWrapper, error) {
	var decoders []*decoderWrapper

	for _, sub := range s.subscriptions {
		for _, src := range sub.Playlists() {
			reader, err := listing.CreateReader(ctx, s.httpClient, src)
			if err != nil {
				s.closeDecoders(decoders)
				return nil, err
			}

			decoder := m3u8.NewDecoder(reader)
			decoders = append(decoders, &decoderWrapper{
				decoder:      decoder,
				subscription: sub,
				reader:       reader,
			})
		}
	}

	return decoders, nil
}

func (s *Streamer) closeDecoders(decoders []*decoderWrapper) {
	for _, d := range decoders {
		d.Close()
	}
}

func (s *Streamer) createRulesProcessor() *rules.Processor {
	processor := rules.NewProcessor()

	for _, sub := range s.subscriptions {
		processor.AddSubscription(sub)
	}

	return processor
}
