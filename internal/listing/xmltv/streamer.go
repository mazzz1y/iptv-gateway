package xmltv

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"iptv-gateway/internal/app"
	"iptv-gateway/internal/ioutil"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/parser/xmltv"
	"iptv-gateway/internal/urlgen"
	"net/http"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultChannelMapSize   = 10000
	DefaultProgrammeMapSize = 100000
)

type Streamer struct {
	subscriptions       []listing.Subscription
	httpClient          listing.HTTPClient
	channels            map[string]bool
	addedChannels       map[string]bool
	addedProgrammes     map[string]bool
	currentURLGenerator listing.URLGenerator
	mu                  sync.Mutex
}

func NewStreamer(subs []*app.Subscription, httpClient listing.HTTPClient, channels map[string]bool) *Streamer {
	subscriptions := make([]listing.Subscription, len(subs))
	for i, sub := range subs {
		subscriptions[i] = sub
	}
	return &Streamer{
		subscriptions:   subscriptions,
		httpClient:      httpClient,
		channels:        channels,
		addedChannels:   make(map[string]bool, DefaultChannelMapSize),
		addedProgrammes: make(map[string]bool, DefaultProgrammeMapSize),
	}
}

func (s *Streamer) WriteToGzip(ctx context.Context, w io.Writer) (int64, error) {
	gzWriter, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	defer gzWriter.Close()

	return s.WriteTo(ctx, gzWriter)
}

func (s *Streamer) WriteTo(ctx context.Context, w io.Writer) (int64, error) {
	if len(s.subscriptions) == 0 {
		return 0, fmt.Errorf("no EPG sources found")
	}

	bytesCounter := ioutil.NewCountWriter(w)
	encoder := xmltv.NewEncoder(bytesCounter)
	defer encoder.Close()

	decoders, err := s.initDecoders(ctx)
	if err != nil {
		return bytesCounter.Count(), err
	}

	defer func() {
		for _, decoder := range decoders {
			decoder.Close()
		}
	}()

	if err := s.processChannels(ctx, decoders, encoder); err != nil {
		return bytesCounter.Count(), err
	}

	if err := s.processProgrammes(ctx, decoders, encoder); err != nil {
		return bytesCounter.Count(), err
	}

	count := bytesCounter.Count()
	if count == 0 {
		return count, fmt.Errorf("no data in subscriptions")
	}

	return count, nil
}

func (s *Streamer) initDecoders(ctx context.Context) ([]*decoderWrapper, error) {
	var decoders []*decoderWrapper

	for _, sub := range s.subscriptions {
		for _, epgURL := range sub.GetEPGs() {
			req, err := http.NewRequestWithContext(ctx, "GET", epgURL, nil)
			if err != nil {
				s.closeDecoders(decoders)
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			resp, err := s.httpClient.Do(req)
			if err != nil {
				s.closeDecoders(decoders)
				return nil, fmt.Errorf("failed to fetch EPG: %w", err)
			}

			decoder := xmltv.NewDecoder(resp.Body)
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
func (s *Streamer) processChannels(ctx context.Context, decoders []*decoderWrapper, encoder xmltv.Encoder) error {
	return listing.Process(
		ctx,
		decoders,
		func(ctx context.Context, decoder *decoderWrapper, output chan<- listing.Item, errChan chan<- error) {
			if decoder.subscription.IsProxied() {
				s.mu.Lock()
				s.currentURLGenerator = decoder.subscription.GetURLGenerator()
				s.mu.Unlock()
			}

			for {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				default:
				}

				item, err := decoder.nextItem()
				if err == io.EOF {
					decoder.done = true
					return
				}
				if err != nil {
					errChan <- err
					return
				}

				switch v := item.(type) {
				case xmltv.Channel:
					output <- v

				case xmltv.Programme:
					decoder.channelsDone = true
					decoder.bufferedItem = v
					decoder.hasBufferedItem = true
					return
				}
			}
		},
		func(item listing.Item) error {
			channel := item.(xmltv.Channel)

			if !s.isChannelAllowed(channel) {
				return nil
			}

			s.mu.Lock()
			if s.currentURLGenerator != nil {
				channel.Icons = s.processIcons(channel.Icons)
			}
			s.mu.Unlock()

			return encoder.Encode(channel)
		},
	)
}

func (s *Streamer) processProgrammes(ctx context.Context, decoders []*decoderWrapper, encoder xmltv.Encoder) error {
	activeDecoders := make([]*decoderWrapper, 0, len(decoders))
	for _, decoder := range decoders {
		if !decoder.done && decoder.err == nil {
			activeDecoders = append(activeDecoders, decoder)
		}
	}

	if len(activeDecoders) == 0 {
		return nil
	}

	return listing.Process(
		ctx,
		activeDecoders,
		func(ctx context.Context, decoder *decoderWrapper, output chan<- listing.Item, errChan chan<- error) {
			if decoder.subscription.IsProxied() {
				s.mu.Lock()
				s.currentURLGenerator = decoder.subscription.GetURLGenerator()
				s.mu.Unlock()
			}

			for {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				default:
				}

				item, err := decoder.nextItem()
				if err == io.EOF {
					return
				}
				if err != nil {
					errChan <- err
					return
				}

				if programme, ok := item.(xmltv.Programme); ok {
					output <- programme
				}
			}
		},
		func(item listing.Item) error {
			programme := item.(xmltv.Programme)

			if !s.isProgrammeAllowed(programme) {
				return nil
			}

			s.mu.Lock()
			if s.currentURLGenerator != nil {
				programme.Icons = s.processIcons(programme.Icons)
			}
			s.mu.Unlock()

			err := encoder.Encode(programme)
			if errors.Is(err, syscall.EPIPE) {
				return nil
			}
			return err
		},
	)
}

func (s *Streamer) isChannelAllowed(channel xmltv.Channel) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

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

	urlData := urlgen.Data{
		RequestType: urlgen.File,
	}

	for i := range result {
		if result[i].Source == "" {
			continue
		}

		urlData.URL = result[i].Source
		link, err := s.currentURLGenerator.CreateURL(urlData, 0)
		if err != nil {
			continue
		}
		result[i].Source = link.String()
	}

	return result
}
