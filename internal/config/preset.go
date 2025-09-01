package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Preset struct {
	Proxy         Proxy              `yaml:"proxy,omitempty"`
	Rules         []rules.RuleAction `yaml:"rules,omitempty"`
	Subscriptions types.StringOrArr  `yaml:"subscriptions"`
}
