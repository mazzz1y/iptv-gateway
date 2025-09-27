package types

import (
	"gopkg.in/yaml.v3"
)

type ConditionList []Condition

type Condition struct {
	NamePatterns RegexpArr     `yaml:"name_patterns,omitempty"`
	Attr         *NamePatterns `yaml:"attr,omitempty"`
	Tag          *NamePatterns `yaml:"tag,omitempty"`
	User         StringOrArr   `yaml:"user,omitempty"`
	Playlist     StringOrArr   `yaml:"playlist,omitempty"`
	Invert       bool          `yaml:"invert,omitempty"`
	And          ConditionList `yaml:"and,omitempty"`
	Or           ConditionList `yaml:"or,omitempty"`
}

func (c *Condition) IsEmpty() bool {
	return c.NamePatterns == nil &&
		c.Attr == nil && c.Tag == nil && len(c.User) == 0 && len(c.Playlist) == 0 &&
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
