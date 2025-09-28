package rules

import (
	"iptv-gateway/internal/config/types"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSetFieldRule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid name template",
			yaml: `
when:
  name_patterns: ["test"]
set_field:
  name: "{{.Name}}"`,
			wantErr: false,
		},
		{
			name: "valid attr template",
			yaml: `
when:
  name_patterns: ["test"]
set_field:
  attr:
    name: tvg-id
    template: "{{.Name}}"`,
			wantErr: false,
		},
		{
			name: "valid tag template",
			yaml: `
when:
  name_patterns: ["test"]
set_field:
  tag:
    name: quality
    template: "{{.Name}}"`,
			wantErr: false,
		},
		{
			name: "missing set_field",
			yaml: `
when:
  name_patterns: ["test"]`,
			wantErr: true,
		},
		{
			name: "multiple templates",
			yaml: `
when:
  name_patterns: ["test"]
set_field:
  name: "{{.Name}}"
  attr:
    name: tvg-id
    template: "{{.Name}}"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rule SetFieldRule
			if err := yaml.Unmarshal([]byte(tt.yaml), &rule); err != nil {
				t.Errorf("failed to unmarshal yaml: %v", err)
				return
			}

			err := rule.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetFieldRule_String(t *testing.T) {
	rule := &SetFieldRule{}
	if rule.String() != "set_field" {
		t.Errorf("String() = %v, want %v", rule.String(), "set_field")
	}
}
