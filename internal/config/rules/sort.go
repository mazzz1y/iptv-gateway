package rules

import (
	"errors"
	"iptv-gateway/internal/config/types"
)

type SortRule struct {
	Attr    string             `yaml:"attr,omitempty"`
	Tag     string             `yaml:"tag,omitempty"`
	Order   *types.StringOrArr `yaml:"order,omitempty"`
	GroupBy *GroupByRule       `yaml:"group_by,omitempty"`
}

func (s *SortRule) Validate() error {
	if s.Attr != "" && s.Tag != "" {
		return errors.New("sort: attr and tag are mutually exclusive")
	}
	return nil
}

func (s *SortRule) String() string {
	return "sort"
}

type GroupByRule struct {
	Tag   string             `yaml:"tag,omitempty"`
	Attr  string             `yaml:"attr,omitempty"`
	Order *types.StringOrArr `yaml:"group_order,omitempty"`
}

func (g *GroupByRule) Validate() error {
	if g.Tag == "" && g.Attr == "" {
		return errors.New("sort: group_by: tag or attr is required")
	}

	if g.Attr != "" && g.Order != nil {
		return errors.New("sort: group_by: attr and order are mutually exclusive")
	}
	return nil
}
