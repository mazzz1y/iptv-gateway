package m3u8

import (
	"fmt"
	"iptv-gateway/internal/client"
	"iptv-gateway/internal/listing/m3u8/channel"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
	"net/url"
	"strings"
	"sync"
)

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

func (p *Processor) Process(registry *channel.Registry) error {
	p.applyRegistryRules(registry)

	filteredChannels := make([]*channel.Channel, 0, registry.Len())

	for _, ch := range registry.All() {
		track := ch.Track()

		if track.IsRemoved {
			continue
		}

		if p.isDuplicate(track) {
			continue
		}

		if shouldSkip := ch.RulesEngine().ProcessTrack(track); shouldSkip {
			continue
		}

		if ch.Subscription().IsProxied() {
			if err := p.processProxyLinks(track, ch.URLGenerator()); err != nil {
				return err
			}
		}

		filteredChannels = append(filteredChannels, ch)
	}

	p.replaceRegistryChannels(registry, filteredChannels)

	return nil
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

func (p *Processor) processProxyLinks(track *m3u8.Track, urlGenerator channel.URLGenerator) error {
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

func (p *Processor) replaceRegistryChannels(registry *channel.Registry, channels []*channel.Channel) {
	registry.Clear()
	for _, ch := range channels {
		registry.Add(ch)
	}
}

func (p *Processor) applyRegistryRules(registry *channel.Registry) {
	subscriptions := make(map[*client.Subscription][]*channel.Channel)

	for _, ch := range registry.All() {
		sub := ch.Subscription().(*client.Subscription)
		subscriptions[sub] = append(subscriptions[sub], ch)
	}

	for sub, channels := range subscriptions {
		sub.ClearChannels()
		for _, ch := range channels {
			sub.AddChannel(ch)
		}
		sub.ApplyRules(registry)
	}
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Host != ""
}
