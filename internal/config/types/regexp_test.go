package types

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRegexpArr_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		wantErr  bool
		validate func(t *testing.T, arr RegexpArr)
	}{
		{
			name:     "single pattern",
			yamlData: `"test.*"`,
			wantErr:  false,
			validate: func(t *testing.T, arr RegexpArr) {
				if len(arr) != 1 {
					t.Errorf("expected 1 pattern, got %d", len(arr))
				}
				if !arr[0].MatchString("testing") {
					t.Error("pattern should match 'testing'")
				}
			},
		},
		{
			name:     "multiple patterns",
			yamlData: `["test.*", "^hello", "world$"]`,
			wantErr:  false,
			validate: func(t *testing.T, arr RegexpArr) {
				if len(arr) != 3 {
					t.Errorf("expected 3 patterns, got %d", len(arr))
				}
				if !arr[0].MatchString("testing") {
					t.Error("first pattern should match 'testing'")
				}
				if !arr[1].MatchString("hello world") {
					t.Error("second pattern should match 'hello world'")
				}
				if !arr[2].MatchString("hello world") {
					t.Error("third pattern should match 'hello world'")
				}
			},
		},
		{
			name:     "empty array",
			yamlData: `[]`,
			wantErr:  false,
			validate: func(t *testing.T, arr RegexpArr) {
				if len(arr) != 0 {
					t.Errorf("expected 0 patterns, got %d", len(arr))
				}
			},
		},
		{
			name:     "invalid regex pattern",
			yamlData: `"[invalid"`,
			wantErr:  true,
			validate: nil,
		},
		{
			name:     "mixed valid and invalid patterns",
			yamlData: `["valid.*", "[invalid"]`,
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr RegexpArr
			err := yaml.Unmarshal([]byte(tt.yamlData), &arr)

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
				tt.validate(t, arr)
			}
		})
	}
}
