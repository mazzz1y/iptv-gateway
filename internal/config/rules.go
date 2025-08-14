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
	Type     string             `yaml:"-"`
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
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("set_field spec must be a mapping")
	}

	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valueNode := value.Content[i+1]

		if keyNode.Value == "" {
			continue
		}

		fieldType := keyNode.Value

		type fieldContent struct {
			Name     string `yaml:"name"`
			Template string `yaml:"template"`
		}

		var content fieldContent
		if err := valueNode.Decode(&content); err != nil {
			return fmt.Errorf("failed to decode %s field content: %w", fieldType, err)
		}

		if fieldType != "name" && content.Name == "" {
			return fmt.Errorf("field 'name' is required for set_field action with type '%s'", fieldType)
		}
		if content.Template == "" {
			return fmt.Errorf("field 'template' is required for set_field action")
		}

		tmpl, err := template.New(fieldType + ":" + content.Name).Funcs(sprig.TxtFuncMap()).Parse(content.Template)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		sfs.Type = fieldType
		sfs.Name = content.Name
		sfs.Template = tmpl

		return nil
	}

	return fmt.Errorf("no field type found in set_field spec")
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
