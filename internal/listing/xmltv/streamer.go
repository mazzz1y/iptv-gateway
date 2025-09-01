package xmltv

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"iptv-gateway/internal/app"
	"iptv-gateway/internal/ioutil"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/parser/xmltv"
	"iptv-gateway/internal/urlgen"
	"net/http"
	"sync"
	"time"
)

type Streamer struct {
	subscriptions    []listing.Subscription
	httpClient       listing.HTTPClient
	channelIDToName  map[string]string
	addedChannelIDs  map[string]bool
	addedProgrammes  map[string]bool
	channelIDMapping map[string]string
	mu               sync.RWMutex
}

func NewStreamer(subs []*app.Subscription, httpClient listing.HTTPClient, channelIDToName map[string]string) *Streamer {
	subscriptions := make([]listing.Subscription, len(subs))
	for i, sub := range subs {
		subscriptions[i] = sub
	}

	channelLen := len(channelIDToName)
	approxProgrammeLen := 300 * channelLen

	return &Streamer{
		subscriptions:    subscriptions,
		httpClient:       httpClient,
		channelIDToName:  channelIDToName,
		channelIDMapping: make(map[string]string, channelLen),
		addedChannelIDs:  make(map[string]bool, channelLen),
		addedProgrammes:  make(map[string]bool, approxProgrammeLen),
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
					v.Icons = s.processIcons(decoder.subscription, v.Icons)
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
			if !s.processChannel(&channel) {
				return nil
			}

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
					programme.Icons = s.processIcons(decoder.subscription, programme.Icons)
					if !s.processProgramme(&programme) {
						continue
					}
					output <- programme
				}
			}
		},
		func(item listing.Item) error {
			return encoder.Encode(item)
		},
	)
}

func (s *Streamer) processChannel(channel *xmltv.Channel) (allowed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalID := channel.ID

	allIDs := []string{originalID}
	for _, displayName := range channel.DisplayNames {
		tvgID := listing.GenerateTvgID(displayName.Value)
		allIDs = append(allIDs, tvgID)
	}

	for _, id := range allIDs {
		if s.addedChannelIDs[id] {
			return false
		}
	}

	for _, id := range allIDs {
		if channelName, exists := s.channelIDToName[id]; exists {
			for _, processedID := range allIDs {
				s.addedChannelIDs[processedID] = true
			}

			if id != originalID {
				s.channelIDMapping[originalID] = id
				channel.ID = id
			}

			if channelName != "" {
				channel.DisplayNames = []xmltv.CommonElement{
					{
						Value: channelName,
					},
				}
			}

			return true
		}
	}

	return false
}

func (s *Streamer) processProgramme(programme *xmltv.Programme) (allowed bool) {
	if mappedChannel, exists := s.channelIDMapping[programme.Channel]; exists {
		programme.Channel = mappedChannel
	}

	if !s.addedChannelIDs[programme.Channel] {
		return false
	}

	key := programme.Channel
	if programme.Start != nil {
		key += programme.Start.Time.Format(time.RFC3339)
	}
	if programme.ID != "" {
		key += programme.ID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.addedProgrammes[key] {
		return false
	}

	s.addedProgrammes[key] = true
	return true
}

func (s *Streamer) processIcons(sub listing.Subscription, icons []xmltv.Icon) []xmltv.Icon {
	gen := sub.GetURLGenerator()
	if gen == nil || len(icons) == 0 {
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

	urlData := urlgen.Data{
		RequestType: urlgen.File,
	}

	for i := range icons {
		if icons[i].Source == "" {
			continue
		}

		urlData.URL = icons[i].Source
		link, err := gen.CreateURL(urlData, 0)
		if err != nil {
			continue
		}
		icons[i].Source = link.String()
	}

	return icons
}
