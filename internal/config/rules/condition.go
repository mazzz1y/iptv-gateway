package rules

import "iptv-gateway/internal/config/types"

type Condition struct {
	Name types.RegexpArr     `yaml:"name,omitempty"`
	Attr *AttributeCondition `yaml:"attr,omitempty"`
	Tag  *TagCondition       `yaml:"tag,omitempty"`
	And  []Condition         `yaml:"and,omitempty"`
	Or   []Condition         `yaml:"or,omitempty"`
	Not  []Condition         `yaml:"not,omitempty"`
}

type AttributeCondition struct {
	Name  string          `yaml:"name"`
	Value types.RegexpArr `yaml:"value"`
}

type TagCondition struct {
	Name  string          `yaml:"name"`
	Value types.RegexpArr `yaml:"value"`
}

func (c *Condition) IsEmpty() bool {
	return len(c.Name) == 0 && c.Attr == nil && c.Tag == nil && len(c.And) == 0 && len(c.Or) == 0 && len(c.Not) == 0
}
