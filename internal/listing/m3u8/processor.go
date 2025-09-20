package m3u8

import (
	"fmt"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
	"net/url"
	"strings"
	"sync"
)

type Processor struct {
	addedTrackIDs   map[string]struct{}
	addedTrackNames map[string]struct{}
	mu              sync.RWMutex
}

func NewProcessor() *Processor {
	return &Processor{
		addedTrackIDs:   make(map[string]struct{}),
		addedTrackNames: make(map[string]struct{}),
	}
}

func (p *Processor) Process(store *rules.Store, rulesProcessor *rules.Processor) ([]*rules.Channel, error) {
	rulesProcessor.Process(store)

	filteredChannels := make([]*rules.Channel, 0, store.Len())

	for _, ch := range store.All() {
		if ch.IsRemoved() {
			continue
		}

		p.ensureTvgID(ch)

		if p.isDuplicate(ch) {
			continue
		}

		if ch.Subscription().IsProxied() {
			if err := p.processProxyLinks(ch); err != nil {
				return nil, err
			}
		}

		filteredChannels = append(filteredChannels, ch)
	}

	return filteredChannels, nil
}

func (p *Processor) ensureTvgID(ch *rules.Channel) {
	if tvgID, exists := ch.GetAttr("tvg-id"); exists && tvgID != "" {
		return
	}
	if tvgName, exists := ch.GetAttr("tvg-name"); exists && tvgName != "" {
		ch.SetAttr("tvg-id", listing.GenerateHashID(tvgName))
	} else {
		ch.SetAttr("tvg-id", listing.GenerateHashID(ch.Name()))
	}
}

func (p *Processor) isDuplicate(ch *rules.Channel) bool {
	id, hasID := ch.GetAttr(m3u8.AttrTvgID)
	trackName := strings.ToLower(ch.Name())

	p.mu.RLock()
	if hasID && id != "" {
		if _, exists := p.addedTrackIDs[id]; exists {
			p.mu.RUnlock()
			return true
		}
	} else {
		if _, exists := p.addedTrackNames[trackName]; exists {
			p.mu.RUnlock()
			return true
		}
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	if hasID && id != "" {
		if _, exists := p.addedTrackIDs[id]; exists {
			return true
		}
		p.addedTrackIDs[id] = struct{}{}
		return false
	}

	if _, exists := p.addedTrackNames[trackName]; exists {
		return true
	}
	p.addedTrackNames[trackName] = struct{}{}
	return false
}

func (p *Processor) processProxyLinks(ch *rules.Channel) error {
	subscription := ch.Subscription()
	urlGen := subscription.URLGenerator()

	for key, value := range ch.Attrs() {
		if isURL(value) {
			encURL, err := urlGen.CreateURL(urlgen.Data{
				RequestType: urlgen.File,
				URL:         value,
			})
			if err != nil {
				return fmt.Errorf("failed to encode attribute URL: %w", err)
			}
			ch.SetAttr(key, encURL.String())
		}
	}

	if ch.URI() != nil {
		uriStr := ch.URI().String()
		if isURL(uriStr) {
			newURL, err := urlGen.CreateURL(urlgen.Data{
				RequestType: urlgen.Stream,
				ChannelID:   ch.Name(),
				URL:         uriStr,
				Hidden:      ch.IsHidden(),
			})
			if err != nil {
				return fmt.Errorf("failed to encode stream URL: %w", err)
			}
			ch.SetURI(newURL)
		}
	}

	return nil
}

func isURL(str string) bool {
	if str == "" {
		return false
	}

	u, err := url.Parse(str)
	return err == nil && u.Host != ""
}
