package types

import "fmt"

type SetFieldTemplate struct {
	NameTemplate *Template     `yaml:"name_template,omitempty"`
	AttrTemplate *NameTemplate `yaml:"attr,omitempty"`
	TagTemplate  *NameTemplate `yaml:"tag,omitempty"`
}

func (s *SetFieldTemplate) Validate() error {
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
		return fmt.Errorf("set_field: exactly one of name_template, attr, or tag is required")
	}

	return nil
}
