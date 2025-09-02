package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Subscription struct {
	Name          string               `yaml:"name"`
	Playlist      types.StringOrArr    `yaml:"playlist_sources"`
	EPG           types.StringOrArr    `yaml:"epg_sources"`
	Proxy         Proxy                `yaml:"proxy"`
	ChannelRules  []rules.ChannelRule  `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule `yaml:"playlist_rules,omitempty"`
}
