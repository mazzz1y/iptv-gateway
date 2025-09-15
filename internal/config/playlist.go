package config

import (
	"fmt"
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/rules/channel"
	"iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/config/types"
)

type Playlist struct {
	Name          string            `yaml:"name"`
	Sources       types.StringOrArr `yaml:"sources"`
	Proxy         proxy.Proxy       `yaml:"proxy,omitempty"`
	ChannelRules  []channel.Rule    `yaml:"channel_rules,omitempty"`
	PlaylistRules []playlist.Rule   `yaml:"playlist_rules,omitempty"`
}

func (p *Playlist) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("playlist name is required")
	}
	if len(p.Sources) == 0 {
		return fmt.Errorf("playlist sources are required")
	}
	for i, source := range p.Sources {
		if source == "" {
			return fmt.Errorf("playlist source[%d] cannot be empty", i)
		}
	}

	for i, rule := range p.ChannelRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("playlist channel_rules[%d]: %w", i, err)
		}
	}

	for i, rule := range p.PlaylistRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("playlist playlist_rules[%d]: %w", i, err)
		}
	}

	return nil
}
