package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Client struct {
	Secret        string             `yaml:"secret"`
	Subscriptions types.StringOrArr  `yaml:"subscriptions"`
	Preset        types.StringOrArr  `yaml:"presets,omitempty"`
	Proxy         Proxy              `yaml:"proxy,omitempty"`
	Rules         []rules.RuleAction `yaml:"rule_list,omitempty"`
}
