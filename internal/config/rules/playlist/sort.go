package playlist

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type Sort struct {
	Selector  *common.Selector  `yaml:"selector,omitempty"`
	Order     *common.RegexpArr `yaml:"order,omitempty"`
	GroupBy   *GroupByRule      `yaml:"group_by,omitempty"`
	Condition *common.Condition `yaml:"condition,omitempty"`
}

func (s *Sort) Validate() error {
	if s.Selector != nil {
		if err := s.Selector.Validate(); err != nil {
			return fmt.Errorf("sort: %w", err)
		}
	}

	if s.Condition != nil {
		if err := s.Condition.Validate(); err != nil {
			return fmt.Errorf("sort: %w", err)
		}

		if s.Condition.Selector != nil || len(s.Condition.Patterns) > 0 || len(s.Condition.Playlists) > 0 || len(s.Condition.And) > 0 || len(s.Condition.Or) > 0 {
			return fmt.Errorf("sort: only clients field is allowed in condition")
		}
	}

	return nil
}

type GroupByRule struct {
	Selector *common.Selector  `yaml:"selector"`
	Order    *common.RegexpArr `yaml:"group_order,omitempty"`
}

func (g *GroupByRule) Validate() error {
	if g.Selector == nil {
		return fmt.Errorf("sort: group_by: selector is required")
	}

	if g.Selector.Type == common.SelectorName {
		return fmt.Errorf("sort: group_by: name selector is not supported")
	}

	if err := g.Selector.Validate(); err != nil {
		return fmt.Errorf("sort: group_by: %w", err)
	}

	return nil
}
