package channel

import "iptv-gateway/internal/config/rules"

type RemoveChannelRule struct {
	When *rules.Condition `yaml:"when,omitempty"`
}
