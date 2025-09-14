package config

import (
	"fmt"
	"iptv-gateway/internal/config/rules/channel"
	"iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/config/types"
	"strings"
)

type Config struct {
	Server        ServerConfig       `yaml:"server"`
	Logs          Logs               `yaml:"logs"`
	URLGenerator  URLGeneratorConfig `yaml:"url_generator"`
	Cache         CacheConfig        `yaml:"cache"`
	Proxy         Proxy              `yaml:"proxy"`
	Clients       []Client           `yaml:"clients"`
	Playlists     []Playlist         `yaml:"playlists"`
	EPGs          []EPG              `yaml:"epgs"`
	ChannelRules  []channel.Rule     `yaml:"channel_rules,omitempty"`
	PlaylistRules []playlist.Rule    `yaml:"playlist_rules,omitempty"`
	Presets       []Preset           `yaml:"presets,omitempty"`
	YamlAnchors   map[string]any     `yaml:",inline"`
}

func (c *Config) Validate() error {
	for key := range c.YamlAnchors {
		if !strings.HasPrefix(key, ".") {
			return fmt.Errorf("unknown config key: %s", key)
		}
	}
	return nil
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
	Name          string            `yaml:"name"`
	Secret        string            `yaml:"secret"`
	Playlists     types.StringOrArr `yaml:"playlists"`
	EPGs          types.StringOrArr `yaml:"epgs"`
	Presets       types.StringOrArr `yaml:"presets,omitempty"`
	Proxy         Proxy             `yaml:"proxy,omitempty"`
	ChannelRules  []channel.Rule    `yaml:"channel_rules,omitempty"`
	PlaylistRules []playlist.Rule   `yaml:"playlist_rules,omitempty"`
}

type CacheConfig struct {
	Path        string         `yaml:"path"`
	TTL         types.Duration `yaml:"ttl"`
	Retention   types.Duration `yaml:"retention"`
	Compression bool           `yaml:"compression"`
}

type EPG struct {
	Name    string            `yaml:"name"`
	Sources types.StringOrArr `yaml:"sources"`
	Proxy   Proxy             `yaml:"proxy,omitempty"`
}

type Playlist struct {
	Name          string            `yaml:"name"`
	Sources       types.StringOrArr `yaml:"sources"`
	Proxy         Proxy             `yaml:"proxy,omitempty"`
	ChannelRules  []channel.Rule    `yaml:"channel_rules,omitempty"`
	PlaylistRules []playlist.Rule   `yaml:"playlist_rules,omitempty"`
}

type Preset struct {
	Name          string            `yaml:"name"`
	Proxy         Proxy             `yaml:"proxy,omitempty"`
	Playlists     types.StringOrArr `yaml:"playlists"`
	EPGs          types.StringOrArr `yaml:"epgs"`
	ChannelRules  []channel.Rule    `yaml:"channel_rules,omitempty"`
	PlaylistRules []playlist.Rule   `yaml:"playlist_rules,omitempty"`
}
