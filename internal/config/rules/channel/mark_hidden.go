package channel

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type MarkHiddenRule struct {
	Condition *common.Condition `yaml:"condition,omitempty"`
}

func (m *MarkHiddenRule) Validate() error {
	if m.Condition == nil {
		return fmt.Errorf("mark_hidden: condition is required")
	}
	if err := m.Condition.Validate(); err != nil {
		return fmt.Errorf("mark_hidden: %w", err)
	}
	return nil
}
