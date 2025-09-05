package rules

import (
	"iptv-gateway/internal/config/types"

	"gopkg.in/yaml.v3"
)

type NamedCondition struct {
	Name string      `yaml:"name"`
	When []Condition `yaml:"when"`
}

type ConditionList []Condition

type Condition struct {
	Name types.RegexpArr     `yaml:"name,omitempty"`
	Attr *AttributeCondition `yaml:"attr,omitempty"`
	Tag  *TagCondition       `yaml:"tag,omitempty"`
	And  ConditionList       `yaml:"and,omitempty"`
	Or   ConditionList       `yaml:"or,omitempty"`
	Not  ConditionList       `yaml:"not,omitempty"`
	Ref  string              `yaml:",omitempty"`
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
	return len(c.Name) == 0 &&
		c.Ref == "" && c.Attr == nil && c.Tag == nil && len(c.And) == 0 && len(c.Or) == 0 && len(c.Not) == 0
}

func (cl *ConditionList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		condition := Condition{Ref: value.Value}
		*cl = ConditionList{condition}
		return nil
	}

	if value.Kind == yaml.SequenceNode {
		var conditions []Condition
		if err := value.Decode(&conditions); err != nil {
			return err
		}
		*cl = conditions
		return nil
	}

	return nil
}

func (c *Condition) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		c.Ref = value.Value
		return nil
	}

	type condition Condition
	var temp condition
	if err := value.Decode(&temp); err != nil {
		return err
	}
	*c = Condition(temp)
	return nil
}
