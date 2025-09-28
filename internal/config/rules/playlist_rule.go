package rules

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type PlaylistRule struct {
	Validate func() error

	RemoveDuplicates *RemoveDuplicatesRule `yaml:"remove_duplicates,omitempty"`
	MergeChannels    *MergeChannelsRule    `yaml:"merge_channels,omitempty"`
	SortRule         *SortRule             `yaml:"sort,omitempty"`
}

func (r *PlaylistRule) UnmarshalYAML(value *yaml.Node) error {
	type rawRule PlaylistRule
	var rr rawRule

	if err := value.Decode(&rr); err != nil {
		return err
	}

	rule := PlaylistRule(rr)

	switch {
	case rule.RemoveDuplicates != nil:
		rule.Validate = rule.RemoveDuplicates.Validate
	case rule.MergeChannels != nil:
		rule.Validate = rule.MergeChannels.Validate
	case rule.SortRule != nil:
		rule.Validate = rule.SortRule.Validate
	default:
		return fmt.Errorf("exactly one playlist rule type must be specified")
	}

	*r = rule
	return nil
}
