package types

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPublicURL_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid http URL",
			yamlData: `"http://example.com"`,
			expected: "http://example.com",
			wantErr:  false,
		},
		{
			name:     "valid https URL",
			yamlData: `"https://example.com:8080/path"`,
			expected: "https://example.com:8080/path",
			wantErr:  false,
		},
		{
			name:     "URL with query params",
			yamlData: `"https://example.com/path?param=value"`,
			expected: "https://example.com/path?param=value",
			wantErr:  false,
		},
		{
			name:     "invalid URL - no host",
			yamlData: `"/path/only"`,
			wantErr:  true,
		},
		{
			name:     "invalid URL - malformed",
			yamlData: `"://invalid"`,
			wantErr:  true,
		},
		{
			name:     "empty URL",
			yamlData: `""`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u PublicURL
			err := yaml.Unmarshal([]byte(tt.yamlData), &u)

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

			if u.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, u.String())
			}
		})
	}
}
