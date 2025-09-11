package rules

import (
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

type SetFieldSpec struct {
	Type  string             `yaml:"-"`
	Name  string             `yaml:"name,omitempty"`
	Value *template.Template `yaml:"-"`
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

		if valueNode.Kind == yaml.ScalarNode {
			templateStr := valueNode.Value

			tmpl, err := template.New("name").Funcs(sprig.TxtFuncMap()).Parse(templateStr)
			if err != nil {
				return fmt.Errorf("failed to parse template: %w", err)
			}

			sfs.Type = "name"
			sfs.Name = ""
			sfs.Value = tmpl

			return nil
		}

		type fieldContent struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		}

		var content fieldContent
		if err := valueNode.Decode(&content); err != nil {
			return fmt.Errorf("failed to decode %s field content: %w", fieldType, err)
		}

		if fieldType != "name" && content.Name == "" {
			return fmt.Errorf("field 'name' is required for set_field action with type '%s'", fieldType)
		}
		if content.Value == "" {
			return fmt.Errorf("field 'value' is required for set_field action")
		}

		tmpl, err := template.New(fieldType + ":" + content.Name).Funcs(sprig.TxtFuncMap()).Parse(content.Value)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		sfs.Type = fieldType
		sfs.Name = content.Name
		sfs.Value = tmpl

		return nil
	}

	return fmt.Errorf("no field type found in set_field spec")
}
