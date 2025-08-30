package rules

import "iptv-gateway/internal/config/types"

type RuleAction struct {
	When              []Condition                  `yaml:"when,omitempty"`
	RemoveChannel     *RemoveChannelRule           `yaml:"remove_channel,omitempty"`
	RemoveChannelDups *RemoveChannelDupsRule       `yaml:"remove_channel_dups,omitempty"`
	RemoveField       []map[string]types.RegexpArr `yaml:"remove_field,omitempty"`
	SetField          []SetFieldSpec               `yaml:"set_field,omitempty"`
}
