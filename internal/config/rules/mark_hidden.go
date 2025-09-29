package rules

import (
	"errors"
	"iptv-gateway/internal/config/common"
)

type MarkHiddenRule struct {
	Condition *common.Condition `yaml:"condition,omitempty"`
}

func (m *MarkHiddenRule) Validate() error {
	if m.Condition == nil {
		return errors.New("condition is required")
	}
	return m.Condition.Validate()
}

func (m *MarkHiddenRule) String() string {
	return "mark_hidden"
}
