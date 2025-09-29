package rules

import (
	"errors"
	"iptv-gateway/internal/config/common"
)

type SortRule struct {
	Selector *common.Selector    `yaml:"selector,omitempty"`
	Order    *common.StringOrArr `yaml:"order,omitempty"`
	GroupBy  *GroupByRule        `yaml:"group_by,omitempty"`
}

func (s *SortRule) Validate() error {
	if s.Selector != nil {
		if err := s.Selector.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SortRule) String() string {
	return "sort"
}

type GroupByRule struct {
	Selector *common.Selector    `yaml:"selector"`
	Order    *common.StringOrArr `yaml:"group_order,omitempty"`
}

func (g *GroupByRule) Validate() error {
	if g.Selector == nil {
		return errors.New("sort: group_by: selector is required")
	}

	if err := g.Selector.Validate(); err != nil {
		return err
	}

	if g.Selector.Type == common.SelectorAttr && g.Order != nil {
		return errors.New("sort: group_by: attr selector and order are mutually exclusive")
	}
	return nil
}
