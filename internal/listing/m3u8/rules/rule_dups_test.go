package rules

import (
	configrules "iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/parser/m3u8"
	"regexp"
	"testing"
)

func TestRemoveDuplicatesProcessor_extractKey(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		AttrPatterns: &types.NamePatterns{
			Name: "x-tvg-name",
			Patterns: types.RegexpArr{
				regexp.MustCompile(`\[HD\]`),
				regexp.MustCompile(`\(FHD\)`),
				regexp.MustCompile(`HD`),
			},
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pattern at start",
			input:    "HD Channel FieldName",
			expected: "Channel FieldName",
		},
		{
			name:     "pattern at end",
			input:    "Channel FieldName HD",
			expected: "Channel FieldName",
		},
		{
			name:     "pattern in middle",
			input:    "Channel HD FieldName",
			expected: "Channel FieldName",
		},
		{
			name:     "multiple patterns",
			input:    "[HD] Channel (FHD) FieldName HD",
			expected: "Channel FieldName",
		},
		{
			name:     "multiple spaces",
			input:    "Channel    FieldName    With    Spaces",
			expected: "Channel FieldName With Spaces",
		},
		{
			name:     "leading and trailing spaces",
			input:    "   Channel FieldName   ",
			expected: "Channel FieldName",
		},
		{
			name:     "pattern creates double spaces",
			input:    "Channel[HD]FieldName",
			expected: "ChannelFieldName",
		},
		{
			name:     "pattern with spaces creates multiple spaces",
			input:    "Channel [HD] FieldName",
			expected: "Channel FieldName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := &Channel{
				track: &m3u8.Track{
					Name: tt.input,
				},
			}
			result := processor.extractBaseName(ch)
			if result != tt.expected {
				t.Errorf("extractKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveDuplicatesProcessor_extractKey_attr(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		AttrPatterns: &types.NamePatterns{
			Name: "x-tvg-name",
			Patterns: types.RegexpArr{
				regexp.MustCompile(`\+3 \(Омск\)`),
				regexp.MustCompile(`\+3`),
				regexp.MustCompile(`\+7 \(Москва\)`),
			},
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)

	tests := []struct {
		name     string
		attrName string
		attrVal  string
		expected string
	}{
		{
			name:     "matches first pattern",
			attrName: "x-tvg-name",
			attrVal:  "+3 (Омск)",
			expected: "",
		},
		{
			name:     "matches second pattern",
			attrName: "x-tvg-name",
			attrVal:  "+3",
			expected: "",
		},
		{
			name:     "matches third pattern",
			attrName: "x-tvg-name",
			attrVal:  "+7 (Москва)",
			expected: "",
		},
		{
			name:     "no match",
			attrName: "x-tvg-name",
			attrVal:  "Channel FieldName",
			expected: "Channel FieldName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := &Channel{
				track: &m3u8.Track{
					Name:  "Test Channel",
					Attrs: map[string]string{tt.attrName: tt.attrVal},
				},
			}
			result := processor.extractBaseName(ch)
			if result != tt.expected {
				t.Errorf("extractKey(attr=%q) = %q, want %q", tt.attrVal, result, tt.expected)
			}
		})
	}
}
