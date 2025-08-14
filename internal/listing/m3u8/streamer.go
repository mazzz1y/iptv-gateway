package m3u8

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iptv-gateway/internal/ioutil"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/manager"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
	"net/url"
	"strings"
	"sync"
	"syscall"
)

type Streamer struct {
	subscriptions   []*manager.Subscription
	httpClient      listing.HTTPClient
	epgLink         string
	addedTrackIDs   map[string]bool
	addedTrackNames map[string]bool
	mu              sync.Mutex
}

type trackWithSubscription struct {
	track        *m3u8.Track
	subscription *manager.Subscription
}

func NewStreamer(subs []*manager.Subscription, epgLink string, httpClient listing.HTTPClient) *Streamer {
	return &Streamer{
		subscriptions:   subs,
		httpClient:      httpClient,
		epgLink:         epgLink,
		addedTrackIDs:   make(map[string]bool),
		addedTrackNames: make(map[string]bool),
	}
}

func (s *Streamer) initializeDecoders(ctx context.Context) ([]*decoderWrapper, error) {
	var decoders []*decoderWrapper

	for _, sub := range s.subscriptions {
		playlists := sub.GetPlaylists()
		for _, playlistURL := range playlists {
			reader, err := s.httpClient.NewReader(ctx, playlistURL)
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

func (s *Streamer) WriteTo(ctx context.Context, w io.Writer) (int64, error) {
	if len(s.subscriptions) == 0 {
		return 0, fmt.Errorf("no subscriptions found")
	}

	bytesCounter := ioutil.NewCountWriter(w)
	encoder := m3u8.NewEncoder(bytesCounter, map[string]string{"x-tvg-url": s.epgLink})
	defer encoder.Close()

	decoders, err := s.initializeDecoders(ctx)
	if err != nil {
		return bytesCounter.Count(), err
	}

	defer func() {
		for _, decoder := range decoders {
			decoder.Close()
		}
	}()

	if err := s.processTracks(ctx, decoders, encoder); err != nil {
		return bytesCounter.Count(), err
	}

	count := bytesCounter.Count()
	if count == 0 {
		return count, fmt.Errorf("no data in subscriptions")
	}

	return count, nil
}

func (s *Streamer) GetAllChannels(ctx context.Context) (map[string]bool, error) {
	if len(s.subscriptions) == 0 {
		return nil, fmt.Errorf("no subscriptions found")
	}

	channels := make(map[string]bool)
	channelsMu := sync.Mutex{}

	decoders, err := s.initializeDecoders(ctx)
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
					output <- trackWithSubscription{
						track:        track,
						subscription: decoder.subscription,
					}
				}
			}
		},
		func(item listing.Item) error {
			trackSub := item.(trackWithSubscription)
			channelsMu.Lock()
			defer channelsMu.Unlock()

			if id, hasID := trackSub.track.Attrs[m3u8.AttrTvgID]; hasID && id != "" {
				channels[id] = true
			}
			channels[trackSub.track.Name] = true
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels found in subscriptions")
	}

	return channels, nil
}

func (s *Streamer) processTracks(ctx context.Context, decoders []*decoderWrapper, encoder m3u8.Encoder) error {
	return listing.Process(
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
					output <- trackWithSubscription{
						track:        track,
						subscription: decoder.subscription,
					}
				}
			}
		},
		func(item listing.Item) error {
			trackSub := item.(trackWithSubscription)

			if s.isDuplicate(trackSub.track) {
				return nil
			}

			if shouldSkip := s.applyRules(trackSub.track, trackSub.subscription); shouldSkip {
				return nil
			}

			if trackSub.subscription.IsProxied() {
				urlGenerator := trackSub.subscription.GetURLGenerator().(*urlgen.Generator)
				if err := s.processProxyLinks(trackSub.track, urlGenerator); err != nil {
					return err
				}
			}

			err := encoder.Encode(trackSub.track)
			if errors.Is(err, syscall.EPIPE) {
				return nil
			}
			return err
		},
	)
}

func (s *Streamer) applyRules(track *m3u8.Track, subscription *manager.Subscription) bool {
	rul := subscription.GetRulesEngine()
	return rul.Process(track)
}

func (s *Streamer) isDuplicate(track *m3u8.Track) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *Streamer) processProxyLinks(track *m3u8.Track, urlGenerator *urlgen.Generator) error {
	for key, value := range track.Attrs {
		if isURL(value) {
			encURL, err := urlGenerator.CreateURL(urlgen.Data{
				RequestType: urlgen.File,
				URL:         value,
			})
			if err != nil {
				return fmt.Errorf("failed to encode attribute URL: %w", err)
			}
			track.Attrs[key] = encURL.String()
		}
	}

	if track.URI != nil && isURL(track.URI.String()) {
		newURL, err := urlGenerator.CreateURL(urlgen.Data{
			RequestType: urlgen.Stream,
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
