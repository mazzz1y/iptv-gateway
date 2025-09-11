package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProxy_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected Proxy
	}{
		{
			name:     "boolean true",
			yamlData: `true`,
			expected: Proxy{
				Enabled: boolPtr(true),
			},
		},
		{
			name:     "boolean false",
			yamlData: `false`,
			expected: Proxy{
				Enabled: boolPtr(false),
			},
		},
		{
			name: "full proxy config",
			yamlData: `
enabled: true
concurrency: 10
stream:
  command: ["ffmpeg", "-i", "{{.url}}", "pipe:1"]
error:
  command: ["error-command"]
  rate_limit_exceeded:
    template_vars:
      - name: message
        value: "Rate limited"`,
			expected: Proxy{
				Enabled:           boolPtr(true),
				ConcurrentStreams: 10,
				Stream: Handler{
					Command: []string{"ffmpeg", "-i", "{{.url}}", "pipe:1"},
				},
				Error: Error{
					Handler: Handler{
						Command: []string{"error-command"},
					},
					RateLimitExceeded: Handler{
						TemplateVars: []EnvNameValue{
							{Name: "message", Value: "Rate limited"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var proxy Proxy
			err := yaml.Unmarshal([]byte(tt.yamlData), &proxy)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expected.Enabled != nil {
				if proxy.Enabled == nil {
					t.Errorf("expected Enabled to be %v, got nil", *tt.expected.Enabled)
				} else if *proxy.Enabled != *tt.expected.Enabled {
					t.Errorf("expected Enabled to be %v, got %v", *tt.expected.Enabled, *proxy.Enabled)
				}
			}

			if proxy.ConcurrentStreams != tt.expected.ConcurrentStreams {
				t.Errorf("expected ConcurrentStreams to be %d, got %d", tt.expected.ConcurrentStreams, proxy.ConcurrentStreams)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
