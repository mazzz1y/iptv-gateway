package m3u8

import (
	"context"
	"fmt"
	"io"
	"iptv-gateway/internal/client"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/listing/m3u8/channel"
	"iptv-gateway/internal/parser/m3u8"
)

type PlaylistFetcher struct {
	subscriptions []*client.Subscription
	httpClient    listing.HTTPClient
}

type TrackSubscription struct {
	Track        *m3u8.Track
	Subscription *client.Subscription
}

func NewPlaylistFetcher(subs []*client.Subscription, httpClient listing.HTTPClient) *PlaylistFetcher {
	return &PlaylistFetcher{
		subscriptions: subs,
		httpClient:    httpClient,
	}
}

func (f *PlaylistFetcher) Fetch(ctx context.Context) (*channel.Registry, error) {
	if len(f.subscriptions) == 0 {
		return nil, fmt.Errorf("no subscriptions found")
	}

	registry := channel.NewRegistry()

	decoders, err := f.initDecoders(ctx)
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
					decoder.done = true
					return
				}
				if err != nil {
					errChan <- err
					return
				}

				if track, ok := item.(*m3u8.Track); ok {
					output <- TrackSubscription{
						Track:        track,
						Subscription: decoder.subscription,
					}
				}
			}
		},
		func(item listing.Item) error {
			trackSub := item.(TrackSubscription)
			ch := channel.NewWithDependencies(
				trackSub.Track,
				trackSub.Subscription,
				trackSub.Subscription.GetURLGenerator(),
				trackSub.Subscription.GetRulesEngine(),
			)
			registry.Add(ch)
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if registry.Len() == 0 {
		return nil, fmt.Errorf("no channels found in subscriptions")
	}

	return registry, nil
}

func (f *PlaylistFetcher) initDecoders(ctx context.Context) ([]*decoderWrapper, error) {
	var decoders []*decoderWrapper

	for _, sub := range f.subscriptions {
		playlists := sub.GetPlaylists()
		for _, playlistURL := range playlists {
			reader, err := f.httpClient.NewReader(ctx, playlistURL)
			if err != nil {
				for _, d := range decoders {
					d.Close()
				}
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
