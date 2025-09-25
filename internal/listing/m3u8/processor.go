package m3u8

import (
	"fmt"
	"iptv-gateway/internal/listing"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
	"net/url"
	"strings"
)

type Processor struct {
	addedTrackIDs   map[string]*rules.Channel
	addedTrackNames map[string]*rules.Channel
	channelStreams  map[*rules.Channel][]urlgen.Stream
}

func NewProcessor() *Processor {
	return &Processor{
		addedTrackIDs:   make(map[string]*rules.Channel),
		addedTrackNames: make(map[string]*rules.Channel),
		channelStreams:  make(map[*rules.Channel][]urlgen.Stream),
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

		if existingChannel := p.isDuplicate(ch); existingChannel != nil {
			p.addURLChannel(existingChannel, ch)
			continue
		}

		if ch.Subscription().IsProxied() {
			if err := p.processProxyLinks(ch); err != nil {
				return nil, err
			}
		}

		p.trackChannel(ch)
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

func (p *Processor) isDuplicate(ch *rules.Channel) *rules.Channel {
	id, hasID := ch.GetAttr(m3u8.AttrTvgID)
	trackName := strings.ToLower(ch.Name())

	if hasID && id != "" {
		if existingChannel, exists := p.addedTrackIDs[id]; exists {
			return existingChannel
		}
	} else {
		if existingChannel, exists := p.addedTrackNames[trackName]; exists {
			return existingChannel
		}
	}

	return nil
}

func (p *Processor) trackChannel(ch *rules.Channel) {
	id, hasID := ch.GetAttr(m3u8.AttrTvgID)
	trackName := strings.ToLower(ch.Name())

	if hasID && id != "" {
		p.addedTrackIDs[id] = ch
	} else {
		p.addedTrackNames[trackName] = ch
	}
}

func (p *Processor) addURLChannel(existingChannel, newChannel *rules.Channel) {
	if !newChannel.Subscription().IsProxied() || newChannel.URI() == nil {
		return
	}

	newStream := urlgen.Stream{
		ProviderInfo: urlgen.ProviderInfo{
			ProviderType: urlgen.ProviderTypePlaylist,
			ProviderName: newChannel.Subscription().Name(),
		},
		URL:    newChannel.URI().String(),
		Hidden: newChannel.IsHidden(),
	}

	p.channelStreams[existingChannel] = append(p.channelStreams[existingChannel], newStream)

	urlGen := existingChannel.Subscription().URLGenerator()
	u, err := urlGen.CreateStreamURL(existingChannel.Name(), p.channelStreams[existingChannel])
	if err == nil {
		existingChannel.SetURI(u)
	}
}

func (p *Processor) processProxyLinks(ch *rules.Channel) error {
	subscription := ch.Subscription()
	urlGen := subscription.URLGenerator()

	for key, value := range ch.Attrs() {
		if isURL(value) {
			u, err := urlGen.CreateFileURL(urlgen.ProviderInfo{
				ProviderType: urlgen.ProviderTypePlaylist,
				ProviderName: ch.Subscription().Name(),
			}, value)
			if err != nil {
				return fmt.Errorf("failed to encode attribute URL: %w", err)
			}
			ch.SetAttr(key, u.String())
		}
	}

	if ch.URI() != nil {
		uriStr := ch.URI().String()
		if isURL(uriStr) {
			stream := urlgen.Stream{
				ProviderInfo: urlgen.ProviderInfo{
					ProviderType: urlgen.ProviderTypePlaylist,
					ProviderName: ch.Subscription().Name(),
				},
				URL:    uriStr,
				Hidden: ch.IsHidden(),
			}

			p.channelStreams[ch] = []urlgen.Stream{stream}

			u, err := urlGen.CreateStreamURL(ch.Name(), p.channelStreams[ch])
			if err != nil {
				return fmt.Errorf("failed to encode stream URL: %w", err)
			}
			ch.SetURI(u)
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
