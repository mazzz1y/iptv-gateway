package rules

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type RemoveFieldRule struct {
	Selector  *common.Selector  `yaml:"selector"`
	Condition *common.Condition `yaml:"condition,omitempty"`
}

func (r *RemoveFieldRule) Validate() error {
	if r.Selector == nil {
		return fmt.Errorf("remove_field: selector is required")
	}

	if err := r.Selector.Validate(); err != nil {
		return err
	}

	if r.Condition != nil {
		if err := r.Condition.Validate(); err != nil {
			return err
		}
	}

	return nil
}
