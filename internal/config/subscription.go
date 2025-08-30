package config

import (
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Subscription struct {
	Playlist types.StringOrArr  `yaml:"playlists"`
	EPG      types.StringOrArr  `yaml:"epgs"`
	Proxy    Proxy              `yaml:"proxy"`
	Rules    []rules.RuleAction `yaml:"rule_list,omitempty"`
}
