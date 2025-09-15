package xmltv

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"iptv-gateway/internal/ioutil"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/parser/xmltv"
	"iptv-gateway/internal/urlgen"
	"sync"
)

type Streamer struct {
	subscriptions    []listing.EPG
	httpClient       listing.HTTPClient
	channelIDToName  map[string]string
	addedChannelIDs  map[string]bool
	addedProgrammes  map[string]bool
	channelIDMapping map[string]string
	mu               sync.RWMutex
}

type Encoder interface {
	Encode(item any) error
	WriteFooter() error
	Close() error
}

func NewStreamer(subs []listing.EPG, httpClient listing.HTTPClient, channelIDToName map[string]string) *Streamer {
	subscriptions := subs
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

	var decoders []*decoderWrapper
	for _, sub := range s.subscriptions {
		for _, url := range sub.EPGs() {
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
		if err := decoder.StartBuffering(ctx); err != nil {
			return bytesCounter.Count(), err
		}
	}

	for _, decoder := range decoders {
		if err := s.processChannels(ctx, decoder, encoder); err != nil {
			return bytesCounter.Count(), err
		}
	}

	for _, decoder := range decoders {
		if err := s.processProgrammes(ctx, decoder, encoder); err != nil {
			return bytesCounter.Count(), err
		}
	}

	count := bytesCounter.Count()
	if count == 0 {
		return count, fmt.Errorf("no data in subscriptions")
	}

	return count, encoder.WriteFooter()
}

func (s *Streamer) processChannels(ctx context.Context, decoder *decoderWrapper, encoder Encoder) error {
	decoder.StopBuffer()
	defer decoder.StartBuffering(ctx)

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
			if item == nil {
				return nil
			}

			if _, ok := item.(xmltv.Programme); ok {
				decoder.AddToBuffer(item)
				return nil
			}

			if channel, ok := item.(xmltv.Channel); ok {
				channel.Icons = s.processIcons(decoder.subscription, channel.Icons)
				if s.processChannel(&channel) {
					if err := encoder.Encode(channel); err != nil {
						return err
					}
				}
			}
		}
	}
}

func (s *Streamer) processProgrammes(ctx context.Context, decoder *decoderWrapper, encoder Encoder) error {
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
			if item == nil {
				return nil
			}

			if programme, ok := item.(xmltv.Programme); ok {
				programme.Icons = s.processIcons(decoder.subscription, programme.Icons)
				if s.processProgramme(&programme) {
					if err := encoder.Encode(programme); err != nil {
						return err
					}
				}
			}
		}
	}
}

func (s *Streamer) processChannel(channel *xmltv.Channel) (allowed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	originalID := channel.ID

	allIDs := make([]string, 0, 1+len(channel.DisplayNames))
	allIDs = append(allIDs, originalID)
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
		key += programme.Start.Time.String()
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

func (s *Streamer) processIcons(sub listing.EPG, icons []xmltv.Icon) []xmltv.Icon {
	if len(icons) == 0 {
		return icons
	}

	gen := sub.URLGenerator()
	if gen == nil {
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
		link, err := gen.CreateURL(urlData)
		if err != nil {
			continue
		}
		icons[i].Source = link.String()
	}

	return icons
}
