package common

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestDuration_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected Duration
		wantErr  bool
	}{
		{
			name:     "seconds",
			yamlData: `30s`,
			expected: Duration(30 * time.Second),
			wantErr:  false,
		},
		{
			name:     "minutes",
			yamlData: `5m`,
			expected: Duration(5 * time.Minute),
			wantErr:  false,
		},
		{
			name:     "hours",
			yamlData: `2h`,
			expected: Duration(2 * time.Hour),
			wantErr:  false,
		},
		{
			name:     "days",
			yamlData: `7d`,
			expected: Duration(7 * 24 * time.Hour),
			wantErr:  false,
		},
		{
			name:     "weeks",
			yamlData: `2w`,
			expected: Duration(2 * 7 * 24 * time.Hour),
			wantErr:  false,
		},
		{
			name:     "months",
			yamlData: `3M`,
			expected: Duration(3 * 30 * 24 * time.Hour),
			wantErr:  false,
		},
		{
			name:     "years",
			yamlData: `1y`,
			expected: Duration(365 * 24 * time.Hour),
			wantErr:  false,
		},
		{
			name:     "invalid format",
			yamlData: `invalid`,
			wantErr:  true,
		},
		{
			name:     "invalid unit",
			yamlData: `30x`,
			wantErr:  true,
		},
		{
			name:     "invalid number",
			yamlData: `abc5s`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := yaml.Unmarshal([]byte(tt.yamlData), &d)

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

			if d != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, d)
			}
		})
	}
}
