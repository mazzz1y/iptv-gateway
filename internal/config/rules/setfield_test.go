package rules

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSetFieldSpec_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		wantErr  bool
		validate func(t *testing.T, spec SetFieldSpec)
	}{
		{
			name: "valid attr field",
			yamlData: `
attr:
  name: "group-title"
  template: "{{.Channel.Name}} Group"`,
			wantErr: false,
			validate: func(t *testing.T, spec SetFieldSpec) {
				if spec.Type != "attr" {
					t.Errorf("expected Type to be 'attr', got '%s'", spec.Type)
				}
				if spec.Name != "group-title" {
					t.Errorf("expected Name to be 'group-title', got '%s'", spec.Name)
				}
				if spec.Template == nil {
					t.Error("expected Template to be set")
				}
			},
		},
		{
			name: "valid tag field",
			yamlData: `
tag:
  name: "custom-tag"
  template: "prefix-{{.Channel.Tags.existing}}"`,
			wantErr: false,
			validate: func(t *testing.T, spec SetFieldSpec) {
				if spec.Type != "tag" {
					t.Errorf("expected Type to be 'tag', got '%s'", spec.Type)
				}
				if spec.Name != "custom-tag" {
					t.Errorf("expected Name to be 'custom-tag', got '%s'", spec.Name)
				}
			},
		},
		{
			name: "valid name field",
			yamlData: `
name:
  template: "{{.Channel.Name}} - Modified"`,
			wantErr: false,
			validate: func(t *testing.T, spec SetFieldSpec) {
				if spec.Type != "name" {
					t.Errorf("expected Type to be 'name', got '%s'", spec.Type)
				}
				if spec.Name != "" {
					t.Errorf("expected Name to be empty for name field, got '%s'", spec.Name)
				}
			},
		},
		{
			name: "missing template",
			yamlData: `
attr:
  name: "group-title"`,
			wantErr: true,
		},
		{
			name: "missing name for non-name field",
			yamlData: `
attr:
  template: "some template"`,
			wantErr: true,
		},
		{
			name: "empty field spec",
			yamlData: `{}`,
			wantErr: true,
		},
		{
			name: "invalid template syntax",
			yamlData: `
attr:
  name: "test"
  template: "{{.Invalid.Template.Syntax"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var spec SetFieldSpec
			err := yaml.Unmarshal([]byte(tt.yamlData), &spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, spec)
			}
		})
	}
}
