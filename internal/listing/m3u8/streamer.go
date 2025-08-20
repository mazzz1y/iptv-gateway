package m3u8

import (
	"context"
	"fmt"
	"io"
	"iptv-gateway/internal/client"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
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
	channels, err := s.getChannels(ctx)
	if err != nil {
		return 0, err
	}

	writer := NewWriter(s.epgLink)
	return writer.WriteChannels(channels, w)
}

func (s *Streamer) GetAllChannels(ctx context.Context) (map[string]bool, error) {
	channels, err := s.getChannels(ctx)
	if err != nil {
		return nil, err
	}

	channelMap := make(map[string]bool)
	for _, ch := range channels {
		track := ch.Track()
		if id := track.Attrs["tvg-id"]; id != "" {
			channelMap[id] = true
		}
		channelMap[track.Name] = true
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

	return processor.Process(store, rulesProcessor, s.subscriptions)
}

func (s *Streamer) fetchPlaylists(ctx context.Context) (*rules.Store, error) {
	if len(s.subscriptions) == 0 {
		return nil, fmt.Errorf("no subscriptions found")
	}

	store := rules.NewStore()

	decoders, err := s.initDecoders()
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
					output <- ch
				}
			}
		},
		func(item listing.Item) error {
			ch := item.(*rules.Channel)
			store.Add(ch)
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if store.Len() == 0 {
		return nil, fmt.Errorf("no channels found in subscriptions")
	}

	return store, nil
}

func (s *Streamer) initDecoders() ([]*decoderWrapper, error) {
	var decoders []*decoderWrapper

	for _, sub := range s.subscriptions {
		playlists := sub.GetPlaylists()
		for _, playlistURL := range playlists {
			resp, err := s.httpClient.Get(playlistURL)
			if err != nil {
				for _, d := range decoders {
					d.Close()
				}
				return nil, err
			}

			decoder := m3u8.NewDecoder(resp.Body)
			decoders = append(decoders, &decoderWrapper{
				decoder:      decoder,
				subscription: sub,
				reader:       resp.Body,
			})
		}
	}

	return decoders, nil
}

func (s *Streamer) createRulesProcessor() *rules.Processor {
	processor := rules.NewProcessor()

	for _, sub := range s.subscriptions {
		if rul := sub.GetRules(); len(rul) > 0 {
			processor.AddSubscriptionRules(sub, rul)
		}
	}

	return processor
}
