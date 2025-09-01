package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Config struct {
	ListenAddr    string                  `yaml:"listen_addr"`
	PublicURL     types.PublicURL         `yaml:"public_url"`
	LogLevel      string                  `yaml:"log_level"`
	Secret        string                  `yaml:"secret"`
	Cache         CacheConfig             `yaml:"cache"`
	Proxy         Proxy                   `yaml:"proxy"`
	MetricsAddr   string                  `yaml:"metrics_addr,omitempty"`
	Clients       map[string]Client       `yaml:"clients"`
	Subscriptions map[string]Subscription `yaml:"subscriptions"`
	Rules         []rules.RuleAction      `yaml:"rules,omitempty"`
	Presets       map[string]Preset       `yaml:"presets,omitempty"`
}
