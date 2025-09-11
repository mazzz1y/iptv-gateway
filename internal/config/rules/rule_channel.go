package rules

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type ChannelRule struct {
	SetField      *SetFieldRule      `yaml:"set_field,omitempty"`
	RemoveField   *RemoveFieldRule   `yaml:"remove_field,omitempty"`
	RemoveChannel *RemoveChannelRule `yaml:"remove_channel,omitempty"`
	MarkHidden    *MarkHiddenRule    `yaml:"mark_hidden,omitempty"`
}

type SetFieldRule struct {
	When         *Condition          `yaml:"when,omitempty"`
	NameTemplate *types.Template     `yaml:"name,omitempty"`
	AttrTemplate *types.NameTemplate `yaml:"attr,omitempty"`
	TagTemplate  *types.NameTemplate `yaml:"tag,omitempty"`
}

type RemoveFieldRule struct {
	When         *Condition      `yaml:"when,omitempty"`
	AttrPatterns types.RegexpArr `yaml:"attr_patterns,omitempty"`
	TagPatterns  types.RegexpArr `yaml:"tag_patterns,omitempty"`
}

type RemoveChannelRule struct {
	When *Condition `yaml:"when,omitempty"`
}

type MarkHiddenRule struct {
	When *Condition `yaml:"when,omitempty"`
}

func (c *ChannelRule) Validate() error {
	ruleCount := 0
	if c.SetField != nil {
		ruleCount++
		if err := c.SetField.Validate(); err != nil {
			return err
		}
	}
	if c.RemoveField != nil {
		ruleCount++
		if err := c.RemoveField.Validate(); err != nil {
			return err
		}
	}
	if c.RemoveChannel != nil {
		ruleCount++
	}
	if c.MarkHidden != nil {
		ruleCount++
	}

	if ruleCount != 1 {
		return fmt.Errorf("channel rule: exactly one rule type must be specified")
	}
	return nil
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

func (r *RemoveFieldRule) Validate() error {
	patternCount := 0
	if len(r.AttrPatterns) > 0 {
		patternCount++
	}
	if len(r.TagPatterns) > 0 {
		patternCount++
	}

	if patternCount != 1 {
		return fmt.Errorf("remove_field: exactly one of attr_patterns or tag_patterns is required")
	}
	return nil
}
