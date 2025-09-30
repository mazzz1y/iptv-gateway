package channel

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type SetFieldRule struct {
	Selector  *common.Selector  `yaml:"selector"`
	Template  *common.Template  `yaml:"template"`
	Condition *common.Condition `yaml:"condition,omitempty"`
}

func (s *SetFieldRule) Validate() error {
	if s.Selector == nil {
		return fmt.Errorf("set_field: selector is required")
	}

	if err := s.Selector.Validate(); err != nil {
		return fmt.Errorf("set_field: %w", err)
	}

	if s.Template == nil {
		return fmt.Errorf("set_field: template is required")
	}

	if s.Condition != nil {
		if err := s.Condition.Validate(); err != nil {
			return fmt.Errorf("set_field: %w", err)
		}
	}

	return nil
}
