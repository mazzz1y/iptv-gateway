package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Playlist struct {
	Name          string               `yaml:"name"`
	Sources       types.StringOrArr    `yaml:"sources"`
	Proxy         Proxy                `yaml:"proxy,omitempty"`
	ChannelRules  []rules.ChannelRule  `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule `yaml:"playlist_rules,omitempty"`
}
