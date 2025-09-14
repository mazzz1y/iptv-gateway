package channel

import "iptv-gateway/internal/config/rules"

type MarkHiddenRule struct {
	When *rules.Condition `yaml:"when,omitempty"`
}
