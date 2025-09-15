package config

import (
	"fmt"
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/rules/channel"
	"iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/config/types"
)

type Client struct {
	Name          string            `yaml:"name"`
	Secret        string            `yaml:"secret"`
	Playlists     types.StringOrArr `yaml:"playlists"`
	EPGs          types.StringOrArr `yaml:"epgs"`
	Presets       types.StringOrArr `yaml:"presets,omitempty"`
	Proxy         proxy.Proxy       `yaml:"proxy,omitempty"`
	ChannelRules  []channel.Rule    `yaml:"channel_rules,omitempty"`
	PlaylistRules []playlist.Rule   `yaml:"playlist_rules,omitempty"`
}

func (c *Client) Validate(playlistNames, epgNames, presetNames map[string]bool) error {
	if c.Name == "" {
		return fmt.Errorf("client name is required")
	}
	if c.Secret == "" {
		return fmt.Errorf("client secret is required")
	}

	for _, p := range c.Playlists {
		if !playlistNames[p] {
			return fmt.Errorf("client references unknown playlist: %s", p)
		}
	}

	for _, epg := range c.EPGs {
		if !epgNames[epg] {
			return fmt.Errorf("client references unknown EPG: %s", epg)
		}
	}

	for _, preset := range c.Presets {
		if !presetNames[preset] {
			return fmt.Errorf("client references unknown preset: %s", preset)
		}
	}

	for i, rule := range c.ChannelRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("client channel_rules[%d]: %w", i, err)
		}
	}

	for i, rule := range c.PlaylistRules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("client playlist_rules[%d]: %w", i, err)
		}
	}

	return nil
}
