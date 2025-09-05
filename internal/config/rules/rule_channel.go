package rules

import "iptv-gateway/internal/config/types"

type ChannelRule struct {
	SetField      []SetFieldSpec               `yaml:"set_field,omitempty"`
	MarkHidden    *any                         `yaml:"mark_hidden,omitempty"`
	RemoveChannel *any                         `yaml:"remove_channel,omitempty"`
	RemoveField   []map[string]types.RegexpArr `yaml:"remove_field,omitempty"`
	When          ConditionList                `yaml:"when,omitempty"`
}
