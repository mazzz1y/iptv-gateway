package types

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ConditionList []Condition

type Condition struct {
	NamePatterns RegexpArr      `yaml:"name_patterns,omitempty"`
	Attr         *NamePatterns  `yaml:"attr,omitempty"`
	Tag          *NamePatterns  `yaml:"tag,omitempty"`
	Clients      StringOrArr    `yaml:"clients,omitempty"`
	Playlists    StringOrArr    `yaml:"playlists,omitempty"`
	Invert       bool           `yaml:"invert,omitempty"`
	And          ConditionList  `yaml:"and,omitempty"`
	Or           ConditionList  `yaml:"or,omitempty"`
	ExtraFields  map[string]any `yaml:",inline"`
}

func (c *Condition) Validate() error {
	fields := make([]string, 0, len(c.ExtraFields))

	for k := range c.ExtraFields {
		fields = append(fields, k)
	}

	if len(fields) > 0 {
		return fmt.Errorf("unknown extra fields: %v", fields)
	}

	for _, cond := range c.And {
		if err := cond.Validate(); err != nil {
			return err
		}
	}
	for _, cond := range c.Or {
		if err := cond.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Condition) IsEmpty() bool {
	return c.NamePatterns == nil &&
		c.Attr == nil && c.Tag == nil && len(c.Clients) == 0 && len(c.Playlists) == 0 &&
		!c.Invert && len(c.And) == 0 && len(c.Or) == 0
}

func (cl *ConditionList) UnmarshalYAML(value *yaml.Node) error {
	var conditions []Condition
	if err := value.Decode(&conditions); err != nil {
		return err
	}
	*cl = conditions
	return nil
}

func (cl *ConditionList) Validate() error {
	for _, c := range *cl {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	return nil
}
