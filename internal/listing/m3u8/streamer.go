package m3u8

import (
	"context"
	"fmt"
	"io"
	"iptv-gateway/internal/app"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
	"net/http"
)

type Streamer struct {
	subscriptions []listing.Subscription
	httpClient    listing.HTTPClient
	epgLink       string
}

func NewStreamer(subs []*app.Subscription, epgLink string, httpClient listing.HTTPClient) *Streamer {
	subscriptions := make([]listing.Subscription, len(subs))
	for i, sub := range subs {
		subscriptions[i] = sub
	}
	return &Streamer{
		subscriptions: subscriptions,
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

func (s *Streamer) initDecoders(ctx context.Context) ([]*decoderWrapper, error) {
	var decoders []*decoderWrapper

	for _, sub := range s.subscriptions {
		for _, playlistURL := range sub.GetPlaylists() {
			req, err := http.NewRequestWithContext(ctx, "GET", playlistURL, nil)
			if err != nil {
				s.closeDecoders(decoders)
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			resp, err := s.httpClient.Do(req)
			if err != nil {
				s.closeDecoders(decoders)
				return nil, fmt.Errorf("failed to fetch playlist: %w", err)
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

func (s *Streamer) closeDecoders(decoders []*decoderWrapper) {
	for _, d := range decoders {
		d.Close()
	}
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
