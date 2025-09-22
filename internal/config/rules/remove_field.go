package rules

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type RemoveFieldRule struct {
	When         *types.Condition `yaml:"when,omitempty"`
	AttrPatterns types.RegexpArr  `yaml:"attr_patterns,omitempty"`
	TagPatterns  types.RegexpArr  `yaml:"tag_patterns,omitempty"`
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
