package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Config struct {
	ListenAddr    string                    `yaml:"listen_addr"`
	PublicURL     types.URL                 `yaml:"public_url"`
	Log           Logs                      `yaml:"log"`
	Secret        string                    `yaml:"secret"`
	Cache         CacheConfig               `yaml:"cache"`
	Proxy         Proxy                     `yaml:"proxy"`
	MetricsAddr   string                    `yaml:"metrics_addr,omitempty"`
	Clients       []Client                  `yaml:"clients"`
	Playlists     []Playlist                `yaml:"playlists"`
	EPGs          []EPG                     `yaml:"epgs"`
	Conditions    []rules.NamedCondition    `yaml:"conditions,omitempty"`
	ChannelRules  []rules.ChannelRule       `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule      `yaml:"playlist_rules,omitempty"`
	Presets       []Preset                  `yaml:"presets,omitempty"`
}
