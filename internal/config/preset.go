package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Preset struct {
	Name          string               `yaml:"name"`
	Proxy         Proxy                `yaml:"proxy,omitempty"`
	Playlists     types.StringOrArr    `yaml:"playlist"`
	EPGs          types.StringOrArr    `yaml:"epg"`
	ChannelRules  []rules.ChannelRule  `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule `yaml:"playlist_rules,omitempty"`
}
