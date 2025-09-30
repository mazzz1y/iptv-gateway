package playlist

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type MergeDuplicatesFinalValue struct {
	Selector *common.Selector `yaml:"selector"`
	Template *common.Template `yaml:"template"`
}

type MergeDuplicatesRule struct {
	Selector   *common.Selector           `yaml:"selector"`
	Patterns   common.RegexpArr           `yaml:"patterns"`
	FinalValue *MergeDuplicatesFinalValue `yaml:"final_value,omitempty"`
	Condition  *common.Condition          `yaml:"condition,omitempty"`
}

func (r *MergeDuplicatesRule) Validate() error {
	if r.Selector != nil {
		if err := r.Selector.Validate(); err != nil {
			return fmt.Errorf("merge_duplicates: %w", err)
		}
	}

	if len(r.Patterns) < 1 {
		return fmt.Errorf("merge_duplicates: at least 1 pattern is required for merging")
	}

	if r.FinalValue != nil {
		if r.FinalValue.Selector == nil {
			return fmt.Errorf("merge_duplicates: final_value selector is required")
		}
		if err := r.FinalValue.Selector.Validate(); err != nil {
			return fmt.Errorf("merge_duplicates: final_value %w", err)
		}
		if r.FinalValue.Template == nil {
			return fmt.Errorf("merge_duplicates: final_value template is required")
		}
	}

	if r.Condition != nil {
		if err := r.Condition.Validate(); err != nil {
			return fmt.Errorf("merge_duplicates: %w", err)
		}

		if r.Condition.Selector != nil ||
			len(r.Condition.Patterns) > 0 ||
			len(r.Condition.Playlists) > 0 ||
			len(r.Condition.And) > 0 ||
			len(r.Condition.Or) > 0 {
			return fmt.Errorf("merge_duplicates: only clients field is allowed in condition")
		}
	}

	return nil
}
