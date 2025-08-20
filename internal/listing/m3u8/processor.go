package m3u8

import (
	"fmt"
	"iptv-gateway/internal/client"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
	"net/url"
	"strings"
	"sync"
	"time"
)

var StreamLinkTTL = time.Hour * 24 * 30

type Processor struct {
	addedTrackIDs   map[string]bool
	addedTrackNames map[string]bool
	mu              sync.Mutex
}

func NewProcessor() *Processor {
	return &Processor{
		addedTrackIDs:   make(map[string]bool),
		addedTrackNames: make(map[string]bool),
	}
}

func (p *Processor) Process(
	store *rules.Store, rulesProcessor *rules.Processor, subscriptions []*client.Subscription) ([]*rules.Channel, error) {
	rulesProcessor.Process(store)

	subscriptionMap := make(map[rules.Subscription]*client.Subscription)
	for _, sub := range subscriptions {
		subscriptionMap[sub] = sub
	}

	filteredChannels := make([]*rules.Channel, 0, store.Len())

	for _, ch := range store.All() {
		track := ch.Track()

		if track.IsRemoved {
			continue
		}

		if p.isDuplicate(track) {
			continue
		}

		if sub, exists := subscriptionMap[ch.Subscription()]; exists {
			if ch.Subscription().IsProxied() {
				if err := p.processProxyLinks(track, sub.GetURLGenerator()); err != nil {
					return nil, err
				}
			}
		}

		filteredChannels = append(filteredChannels, ch)
	}

	return filteredChannels, nil
}

func (p *Processor) isDuplicate(track *m3u8.Track) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	id, hasID := track.Attrs[m3u8.AttrTvgID]
	trackName := strings.ToLower(track.Name)

	if hasID && id != "" {
		if p.addedTrackIDs[id] {
			return true
		}
		p.addedTrackIDs[id] = true
		return false
	}

	if p.addedTrackNames[trackName] {
		return true
	}
	p.addedTrackNames[trackName] = true
	return false
}

func (p *Processor) processProxyLinks(track *m3u8.Track, urlGenerator rules.URLGenerator) error {
	for key, value := range track.Attrs {
		if isURL(value) {
			encURL, err := urlGenerator.CreateURL(urlgen.Data{
				RequestType: urlgen.File,
				URL:         value,
			}, 0)
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
		}, StreamLinkTTL)
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
