package rules

import "iptv-gateway/internal/config/types"

type RemoveChannelRule struct{}

type RemoveChannelDupsRule []struct {
	Patterns    types.RegexpArr `yaml:"patterns"`
	TrimPattern bool            `yaml:"trim_pattern"`
}
