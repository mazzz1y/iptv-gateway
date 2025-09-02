package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Preset struct {
	Name          string               `yaml:"name"`
	Proxy         Proxy                `yaml:"proxy,omitempty"`
	Subscriptions types.StringOrArr    `yaml:"subscriptions"`
	ChannelRules  []rules.ChannelRule  `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule `yaml:"playlist_rules,omitempty"`
}
