package rules

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type SetFieldRule struct {
	When     *types.Condition        `yaml:"when,omitempty"`
	SetField *types.SetFieldTemplate `yaml:"set_field,inline"`
}

func (s *SetFieldRule) Validate() error {
	if s.SetField == nil {
		return fmt.Errorf("set_field: template is required")
	}

	if err := s.SetField.Validate(); err != nil {
		return err
	}

	if s.When == nil {
		return nil
	}

	return s.When.Validate()
}

func (s *SetFieldRule) String() string {
	return "set_field"
}
