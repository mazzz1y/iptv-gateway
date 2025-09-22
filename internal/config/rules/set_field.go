package rules

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type SetFieldRule struct {
	When         *types.Condition    `yaml:"when,omitempty"`
	NameTemplate *types.Template     `yaml:"name,omitempty"`
	AttrTemplate *types.NameTemplate `yaml:"attr,omitempty"`
	TagTemplate  *types.NameTemplate `yaml:"tag,omitempty"`
}

func (s *SetFieldRule) Validate() error {
	setFields := 0
	if s.NameTemplate != nil {
		setFields++
	}
	if s.AttrTemplate != nil {
		setFields++
		if err := s.AttrTemplate.Validate(); err != nil {
			return fmt.Errorf("set_field: attr validation failed: %w", err)
		}
	}
	if s.TagTemplate != nil {
		setFields++
		if err := s.TagTemplate.Validate(); err != nil {
			return fmt.Errorf("set_field: tag validation failed: %w", err)
		}
	}

	if setFields != 1 {
		return fmt.Errorf("set_field: exactly one of name, attr, or tag is required")
	}
	return nil
}

func (s *SetFieldRule) String() string {
	return "set_field"
}
