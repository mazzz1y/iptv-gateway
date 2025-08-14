package config

import (
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

type RuleAction struct {
	When          []Condition        `yaml:"when,omitempty"`
	RemoveChannel *RemoveChannelRule `yaml:"remove_channel,omitempty"`
	RemoveField   []FieldSpec        `yaml:"remove_field,omitempty"`
	SetField      []SetFieldSpec     `yaml:"set_field,omitempty"`
}

type RemoveChannelRule struct{}

type FieldSpec struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}

type SetFieldSpec struct {
	Type     string             `yaml:"type"`
	Name     string             `yaml:"name,omitempty"`
	Template *template.Template `yaml:"-"`
}

type Condition struct {
	Name RegexpArr           `yaml:"name,omitempty"`
	Attr *AttributeCondition `yaml:"attr,omitempty"`
	Tag  *TagCondition       `yaml:"tag,omitempty"`
	And  []Condition         `yaml:"and,omitempty"`
	Or   []Condition         `yaml:"or,omitempty"`
}

type AttributeCondition struct {
	Name  string    `yaml:"name"`
	Value RegexpArr `yaml:"value"`
}

type TagCondition struct {
	Name  string    `yaml:"name"`
	Value RegexpArr `yaml:"value"`
}

func (sfs *SetFieldSpec) UnmarshalYAML(value *yaml.Node) error {
	type plainSpec struct {
		Type     string `yaml:"type"`
		Name     string `yaml:"name"`
		Template string `yaml:"template"`
	}

	var spec plainSpec
	if err := value.Decode(&spec); err != nil {
		return err
	}

	if spec.Type == "" {
		return fmt.Errorf("field 'type' is required for set_field action")
	}
	if spec.Type != "name" && spec.Name == "" {
		return fmt.Errorf("field 'name' is required for set_field action with type '%s'", spec.Type)
	}
	if spec.Template == "" {
		return fmt.Errorf("field 'template' is required for set_field action")
	}

	tmpl, err := template.New(spec.Type + ":" + spec.Name).Funcs(sprig.TxtFuncMap()).Parse(spec.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	sfs.Type = spec.Type
	sfs.Name = spec.Name
	sfs.Template = tmpl

	return nil
}

func (fs *FieldSpec) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err == nil {
		if str == "name" {
			fs.Type = "name"
			fs.Name = ""
			return nil
		}
		return fmt.Errorf("invalid field spec string: '%s' (only 'name' is supported)", str)
	}

	type fieldSpecYAML FieldSpec
	return value.Decode((*fieldSpecYAML)(fs))
}

func (c *Condition) IsEmpty() bool {
	return len(c.Name) == 0 && c.Attr == nil && c.Tag == nil && len(c.And) == 0 && len(c.Or) == 0
}
