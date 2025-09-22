package rules

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type RuleType int

const (
	ChannelRule RuleType = iota
	StoreRule
)

type Rule struct {
	Type     RuleType
	Validate func() error

	SetField      *SetFieldRule      `yaml:"set_field,omitempty"`
	RemoveField   *RemoveFieldRule   `yaml:"remove_field,omitempty"`
	RemoveChannel *RemoveChannelRule `yaml:"remove_channel,omitempty"`
	MarkHidden    *MarkHiddenRule    `yaml:"mark_hidden,omitempty"`

	RemoveDuplicates *RemoveDuplicatesRule `yaml:"remove_duplicates,omitempty"`
	SortRule         *SortRule             `yaml:"sort,omitempty"`
}

func (r *Rule) UnmarshalYAML(value *yaml.Node) error {
	type rawRule Rule
	var rr rawRule

	if err := value.Decode(&rr); err != nil {
		return err
	}

	rule := Rule(rr)

	switch {
	case rule.SetField != nil:
		rule.Type = ChannelRule
		rule.Validate = rule.SetField.Validate

	case rule.RemoveField != nil:
		rule.Type = ChannelRule
		rule.Validate = rule.RemoveField.Validate

	case rule.RemoveChannel != nil:
		rule.Type = ChannelRule
		rule.Validate = rule.RemoveChannel.Validate

	case rule.MarkHidden != nil:
		rule.Type = ChannelRule
		rule.Validate = rule.MarkHidden.Validate

	case rule.RemoveDuplicates != nil:
		rule.Type = StoreRule
		rule.Validate = rule.RemoveDuplicates.Validate

	case rule.SortRule != nil:
		rule.Type = StoreRule
		rule.Validate = rule.SortRule.Validate

	default:
		return fmt.Errorf("exactly one rule type must be specified")
	}

	*r = rule
	return nil
}
