package rules

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type RemoveFieldRule struct {
	Selector  *common.Selector  `yaml:"selector"`
	Patterns  common.RegexpArr  `yaml:"patterns"`
	Condition *common.Condition `yaml:"condition,omitempty"`
}

func (r *RemoveFieldRule) Validate() error {
	if r.Selector == nil {
		return fmt.Errorf("remove_field: selector is required")
	}

	if err := r.Selector.Validate(); err != nil {
		return err
	}

	if len(r.Patterns) == 0 {
		return fmt.Errorf("remove_field: patterns are required")
	}

	if r.Condition != nil {
		if err := r.Condition.Validate(); err != nil {
			return err
		}
	}

	return nil
}
