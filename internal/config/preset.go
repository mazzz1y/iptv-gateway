package config

import (
	"fmt"
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/rules/channel"
	"iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/config/types"
)

type Preset struct {
	Name          string            `yaml:"name"`
	Proxy         proxy.Proxy       `yaml:"proxy,omitempty"`
	Playlists     types.StringOrArr `yaml:"playlists"`
	EPGs          types.StringOrArr `yaml:"epgs"`
	ChannelRules  []channel.Rule    `yaml:"channel_rules,omitempty"`
	PlaylistRules []playlist.Rule   `yaml:"playlist_rules,omitempty"`
}

func (p *Preset) Validate(playlistNames, epgNames map[string]bool) error {
	if p.Name == "" {
		return fmt.Errorf("preset name is required")
	}

	for _, pl := range p.Playlists {
		if !playlistNames[pl] {
			return fmt.Errorf("preset references unknown playlist: %s", pl)
		}
	}

	for _, epg := range p.EPGs {
		if !epgNames[epg] {
			return fmt.Errorf("preset references unknown EPG: %s", epg)
		}
	}

	for i, rule := range p.ChannelRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("preset channel_rules[%d]: %w", i, err)
		}
	}

	for i, rule := range p.PlaylistRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("preset playlist_rules[%d]: %w", i, err)
		}
	}

	return nil
}
