package types

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStringOrArr_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected StringOrArr
		wantErr  bool
	}{
		{
			name:     "single string",
			yamlData: `"single"`,
			expected: StringOrArr{"single"},
			wantErr:  false,
		},
		{
			name:     "string array",
			yamlData: `["first", "second", "third"]`,
			expected: StringOrArr{"first", "second", "third"},
			wantErr:  false,
		},
		{
			name:     "empty array",
			yamlData: `[]`,
			expected: StringOrArr{},
			wantErr:  false,
		},
		{
			name:     "single item array",
			yamlData: `["single"]`,
			expected: StringOrArr{"single"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s StringOrArr
			err := yaml.Unmarshal([]byte(tt.yamlData), &s)

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

			if !reflect.DeepEqual(s, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
