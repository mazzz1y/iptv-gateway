package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Client struct {
	Name          string               `yaml:"name"`
	Secret        string               `yaml:"secret"`
	Playlists     types.StringOrArr    `yaml:"playlist"`
	EPGs          types.StringOrArr    `yaml:"epg"`
	Preset        types.StringOrArr    `yaml:"preset,omitempty"`
	Proxy         Proxy                `yaml:"proxy,omitempty"`
	ChannelRules  []rules.ChannelRule  `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule `yaml:"playlist_rules,omitempty"`
}
