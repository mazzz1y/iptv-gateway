package channel

import (
	"errors"
	"iptv-gateway/internal/config/rules"
)

type MarkHiddenRule struct {
	When *rules.Condition `yaml:"when,omitempty"`
}

func (m *MarkHiddenRule) Validate() error {
	if m.When == nil {
		return errors.New("when is required")
	}
	return nil
}
