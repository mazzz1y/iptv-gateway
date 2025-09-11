package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Config struct {
	Server        ServerConfig           `yaml:"server"`
	Log           Logs                   `yaml:"log"`
	URLGenerator  URLGeneratorConfig     `yaml:"url_generator"`
	Cache         CacheConfig            `yaml:"cache"`
	Proxy         Proxy                  `yaml:"proxy"`
	Clients       []Client               `yaml:"clients"`
	Playlists     []Playlist             `yaml:"playlists"`
	EPGs          []EPG                  `yaml:"epgs"`
	Conditions    []rules.NamedCondition `yaml:"conditions,omitempty"`
	ChannelRules  []rules.ChannelRule    `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule   `yaml:"playlist_rules,omitempty"`
	Presets       []Preset               `yaml:"presets,omitempty"`
}

type URLGeneratorConfig struct {
	Secret    string         `yaml:"secret"`
	StreamTTL types.Duration `yaml:"stream_ttl"`
	FileTTL   types.Duration `yaml:"file_ttl"`
}

type ServerConfig struct {
	ListenAddr  string    `yaml:"listen_addr"`
	MetricsAddr string    `yaml:"metrics_addr"`
	PublicURL   types.URL `yaml:"public_url"`
}

type Logs struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

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

type CacheConfig struct {
	Path        string         `yaml:"path"`
	TTL         types.Duration `yaml:"ttl"`
	Retention   types.Duration `yaml:"retention"`
	Compression bool           `yaml:"compression"`
}

type EPG struct {
	Name    string            `yaml:"name"`
	Sources types.StringOrArr `yaml:"source"`
	Proxy   Proxy             `yaml:"proxy,omitempty"`
}

type Playlist struct {
	Name          string               `yaml:"name"`
	Sources       types.StringOrArr    `yaml:"source"`
	Proxy         Proxy                `yaml:"proxy,omitempty"`
	ChannelRules  []rules.ChannelRule  `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule `yaml:"playlist_rules,omitempty"`
}

type Preset struct {
	Name          string               `yaml:"name"`
	Proxy         Proxy                `yaml:"proxy,omitempty"`
	Playlists     types.StringOrArr    `yaml:"playlist"`
	EPGs          types.StringOrArr    `yaml:"epg"`
	ChannelRules  []rules.ChannelRule  `yaml:"channel_rules,omitempty"`
	PlaylistRules []rules.PlaylistRule `yaml:"playlist_rules,omitempty"`
}
