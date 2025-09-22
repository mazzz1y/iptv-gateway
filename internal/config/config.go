package config

import (
	"fmt"
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/rules"
	"strings"
)

type Config struct {
	YamlSnippets map[string]any     `yaml:",inline"`
	Server       ServerConfig       `yaml:"server"`
	Logs         Logs               `yaml:"logs"`
	URLGenerator URLGeneratorConfig `yaml:"url_generator"`
	Cache        CacheConfig        `yaml:"cache"`
	Proxy        proxy.Proxy        `yaml:"proxy"`
	Clients      []Client           `yaml:"clients"`
	Playlists    []Playlist         `yaml:"playlists"`
	EPGs         []EPG              `yaml:"epgs"`
	Rules        []*rules.Rule      `yaml:"rules,omitempty"`
	Presets      []Preset           `yaml:"presets,omitempty"`
}

func (c *Config) Validate() error {
	for key := range c.YamlSnippets {
		if !strings.HasPrefix(key, ".") {
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server configuration validation failed: %w", err)
	}

	if err := c.Logs.Validate(); err != nil {
		return fmt.Errorf("logs configuration validation failed: %w", err)
	}

	if err := c.URLGenerator.Validate(); err != nil {
		return fmt.Errorf("url_generator configuration validation failed: %w", err)
	}

	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("cache configuration validation failed: %w", err)
	}

	if err := c.Proxy.Validate(); err != nil {
		return fmt.Errorf("proxy configuration validation failed: %w", err)
	}

	playlistNames := make(map[string]bool)
	epgNames := make(map[string]bool)
	presetNames := make(map[string]bool)

	for i, pl := range c.Playlists {
		if err := pl.Validate(); err != nil {
			return fmt.Errorf("playlist[%d] validation failed: %w", i, err)
		}
		if pl.Name != "" {
			if playlistNames[pl.Name] {
				return fmt.Errorf("duplicate playlist name: %s", pl.Name)
			}
			playlistNames[pl.Name] = true
		}
	}

	for i, epg := range c.EPGs {
		if err := epg.Validate(); err != nil {
			return fmt.Errorf("epg[%d] validation failed: %w", i, err)
		}
		if epg.Name != "" {
			if epgNames[epg.Name] {
				return fmt.Errorf("duplicate EPG name: %s", epg.Name)
			}
			epgNames[epg.Name] = true
		}
	}

	for i, preset := range c.Presets {
		if err := preset.Validate(playlistNames, epgNames); err != nil {
			return fmt.Errorf("preset[%d] validation failed: %w", i, err)
		}
		if preset.Name != "" {
			if presetNames[preset.Name] {
				return fmt.Errorf("duplicate preset name: %s", preset.Name)
			}
			presetNames[preset.Name] = true
		}
	}

	clientNames := make(map[string]bool)
	for i, client := range c.Clients {
		if err := client.Validate(playlistNames, epgNames, presetNames); err != nil {
			return fmt.Errorf("client[%d] validation failed: %w", i, err)
		}
		if client.Name != "" {
			if clientNames[client.Name] {
				return fmt.Errorf("duplicate client name: %s", client.Name)
			}
			clientNames[client.Name] = true
		}
	}

	for i, rule := range c.Rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("global rules[%d] validation failed: %w", i, err)
		}
	}

	return nil
}
