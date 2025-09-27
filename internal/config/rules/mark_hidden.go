package rules

import (
	"errors"
	"iptv-gateway/internal/config/types"
)

type MarkHiddenRule struct {
	When *types.Condition `yaml:"when,omitempty"`
}

func (m *MarkHiddenRule) Validate() error {
	if m.When == nil {
		return errors.New("when is required")
	}
	return m.When.Validate()
}

func (m *MarkHiddenRule) String() string {
	return "mark_hidden"
}
