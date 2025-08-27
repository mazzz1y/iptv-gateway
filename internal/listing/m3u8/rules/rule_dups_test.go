package rules

import (
	"regexp"
	"testing"
)

func TestRemoveDuplicatesRule_extractBaseName(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\[HD\]`),
		regexp.MustCompile(`\(FHD\)`),
		regexp.MustCompile(`HD`),
	}

	rule := NewRemoveDuplicatesRule(patterns, false)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pattern at start",
			input:    "HD Channel Name",
			expected: "Channel Name",
		},
		{
			name:     "pattern at end",
			input:    "Channel Name HD",
			expected: "Channel Name",
		},
		{
			name:     "pattern in middle",
			input:    "Channel HD Name",
			expected: "Channel Name",
		},
		{
			name:     "multiple patterns",
			input:    "[HD] Channel (FHD) Name HD",
			expected: "Channel Name",
		},
		{
			name:     "multiple spaces",
			input:    "Channel    Name    With    Spaces",
			expected: "Channel Name With Spaces",
		},
		{
			name:     "leading and trailing spaces",
			input:    "   Channel Name   ",
			expected: "Channel Name",
		},
		{
			name:     "pattern creates double spaces",
			input:    "Channel[HD]Name",
			expected: "ChannelName",
		},
		{
			name:     "pattern with spaces creates multiple spaces",
			input:    "Channel [HD] Name",
			expected: "Channel Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.extractBaseName(tt.input)
			if result != tt.expected {
				t.Errorf("extractBaseName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
