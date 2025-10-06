package m3u8

import (
	"fmt"
	"majmun/internal/listing/m3u8/rules/channel"
	"majmun/internal/listing/m3u8/rules/playlist"
	"majmun/internal/listing/m3u8/store"
	"majmun/internal/parser/m3u8"
	"majmun/internal/urlgen"
	"net/url"
	"strings"
)

type Processor struct {
	channelsByID   map[string]*store.Channel
	channelsByName map[string]*store.Channel
	channelStreams map[*store.Channel][]urlgen.Stream
}

func NewProcessor() *Processor {
	return &Processor{
		channelsByID:   make(map[string]*store.Channel),
		channelsByName: make(map[string]*store.Channel),
		channelStreams: make(map[*store.Channel][]urlgen.Stream),
	}
}

func (p *Processor) Process(
	st *store.Store, channelProcessor *channel.Processor, playlistProcessor *playlist.Processor) ([]*store.Channel, error) {
	channelProcessor.Apply(st)
	playlistProcessor.Apply(st)

	for _, ch := range st.All() {
		if ch.IsRemoved() {
			continue
		}

		if existingChannel := p.isDuplicate(ch); existingChannel != nil {
			p.mergeStream(existingChannel, ch)
			continue
		}

		if ch.Playlist().IsProxied() {
			if err := p.processProxyLinks(ch); err != nil {
				return nil, err
			}
		}

		p.trackChannel(ch)
		p.channelStreams[ch] = nil
	}

	result := make([]*store.Channel, 0, len(p.channelStreams))
	for _, ch := range st.All() {
		if _, exists := p.channelStreams[ch]; exists {
			result = append(result, ch)
		}
	}

	return result, nil
}

func (p *Processor) isDuplicate(ch *store.Channel) *store.Channel {
	id, hasID := ch.GetAttr(m3u8.AttrTvgID)
	trackName := strings.ToLower(ch.Name())

	if hasID && id != "" {
		if existingChannel, exists := p.channelsByID[id]; exists {
			return existingChannel
		}
	} else {
		if existingChannel, exists := p.channelsByName[trackName]; exists {
			return existingChannel
		}
	}

	return nil
}

func (p *Processor) trackChannel(ch *store.Channel) {
	id, hasID := ch.GetAttr(m3u8.AttrTvgID)
	trackName := strings.ToLower(ch.Name())

	if hasID && id != "" {
		p.channelsByID[id] = ch
	} else {
		p.channelsByName[trackName] = ch
	}
}

func (p *Processor) mergeStream(existingChannel, newChannel *store.Channel) {
	if !newChannel.Playlist().IsProxied() || newChannel.URI() == nil {
		return
	}

	if newChannel.Priority() > existingChannel.Priority() {
		p.trackChannel(newChannel)

		if streams, exists := p.channelStreams[existingChannel]; exists {
			p.channelStreams[newChannel] = streams
			delete(p.channelStreams, existingChannel)
		}

		existingChannel = newChannel
	}

	p.channelStreams[existingChannel] = append(p.channelStreams[existingChannel], urlgen.Stream{
		ProviderInfo: urlgen.ProviderInfo{
			ProviderType: urlgen.ProviderTypePlaylist,
			ProviderName: newChannel.Playlist().Name(),
		},
		URL:    newChannel.URI().String(),
		Hidden: newChannel.IsHidden(),
	})

	urlGen := existingChannel.Playlist().URLGenerator()
	if u, err := urlGen.CreateStreamURL(existingChannel.Name(), p.channelStreams[existingChannel]); err == nil {
		existingChannel.SetURI(u)
	}
}

func (p *Processor) processProxyLinks(ch *store.Channel) error {
	subscription := ch.Playlist()
	urlGen := subscription.URLGenerator()

	for key, value := range ch.Attrs() {
		if isURL(value) {
			u, err := urlGen.CreateFileURL(urlgen.ProviderInfo{
				ProviderType: urlgen.ProviderTypePlaylist,
				ProviderName: ch.Playlist().Name(),
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
					ProviderName: ch.Playlist().Name(),
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
