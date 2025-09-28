package rules

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ChannelRule struct {
	Validate func() error

	SetField      *SetFieldRule      `yaml:"set_field,omitempty"`
	RemoveField   *RemoveFieldRule   `yaml:"remove_field,omitempty"`
	RemoveChannel *RemoveChannelRule `yaml:"remove_channel,omitempty"`
	MarkHidden    *MarkHiddenRule    `yaml:"mark_hidden,omitempty"`
}

func (r *ChannelRule) UnmarshalYAML(value *yaml.Node) error {
	type rawRule ChannelRule
	var rr rawRule

	if err := value.Decode(&rr); err != nil {
		return err
	}

	rule := ChannelRule(rr)

	switch {
	case rule.SetField != nil:
		rule.Validate = rule.SetField.Validate
	case rule.RemoveField != nil:
		rule.Validate = rule.RemoveField.Validate
	case rule.RemoveChannel != nil:
		rule.Validate = rule.RemoveChannel.Validate
	case rule.MarkHidden != nil:
		rule.Validate = rule.MarkHidden.Validate
	default:
		return fmt.Errorf("exactly one channel rule type must be specified")
	}

	*r = rule
	return nil
}
