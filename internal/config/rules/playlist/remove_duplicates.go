package playlist

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type RemoveDuplicatesFinalValue struct {
	Selector *common.Selector `yaml:"selector"`
	Template *common.Template `yaml:"template"`
}

type RemoveDuplicatesRule struct {
	Selector   *common.Selector            `yaml:"selector"`
	Patterns   common.RegexpArr            `yaml:"patterns"`
	FinalValue *RemoveDuplicatesFinalValue `yaml:"final_value,omitempty"`
	Condition  *common.Condition           `yaml:"condition,omitempty"`
}

func (r *RemoveDuplicatesRule) Validate() error {
	if r.Selector != nil {
		if err := r.Selector.Validate(); err != nil {
			return fmt.Errorf("remove_duplicates: %w", err)
		}
	}

	if len(r.Patterns) < 2 {
		return fmt.Errorf("remove_duplicates: at least 2 patterns are required for duplicate detection")
	}

	if r.FinalValue != nil {
		if r.FinalValue.Selector == nil {
			return fmt.Errorf("remove_duplicates: final_value selector is required")
		}
		if err := r.FinalValue.Selector.Validate(); err != nil {
			return fmt.Errorf("remove_duplicates: final_value %w", err)
		}
		if r.FinalValue.Template == nil {
			return fmt.Errorf("remove_duplicates: final_value template is required")
		}
	}

	if r.Condition != nil {
		if err := r.Condition.Validate(); err != nil {
			return fmt.Errorf("remove_duplicates: %w", err)
		}

		if r.Condition.Selector != nil || len(r.Condition.Patterns) > 0 || len(r.Condition.Playlists) > 0 || len(r.Condition.And) > 0 || len(r.Condition.Or) > 0 {
			return fmt.Errorf("remove_duplicates: only clients field is allowed in condition")
		}
	}

	return nil
}
