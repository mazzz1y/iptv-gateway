package pattern_matcher

import (
	"majmun/internal/config/common"
	"majmun/internal/listing/m3u8/store"
	"majmun/internal/parser/m3u8"
	"net/url"
	"regexp"
	"testing"
)

func mustParseURL(rawURL string) *url.URL {
	u, _ := url.Parse(rawURL)
	return u
}

func createChannel(name, tvgID, group string) *store.Channel {
	track := &m3u8.Track{
		Name: name,
		URI:  mustParseURL("http://example.com/stream"),
		Attrs: map[string]string{
			m3u8.AttrTvgID: tvgID,
		},
		Tags: map[string]string{
			m3u8.TagGroup: group,
		},
	}
	return store.NewChannel(track, nil)
}

func createPatterns(patterns ...string) common.RegexpArr {
	var regexps common.RegexpArr
	for _, p := range patterns {
		regexps = append(regexps, regexp.MustCompile(p))
	}
	return regexps
}

func TestPatternMatcher_GroupChannels(t *testing.T) {
	tests := []struct {
		name           string
		channels       []*store.Channel
		selector       *common.Selector
		patterns       common.RegexpArr
		expectedGroups map[string][]string // baseName -> channel names in order
	}{
		{
			name: "no patterns, no grouping",
			channels: []*store.Channel{
				createChannel("Channel1", "ch1", "Sports"),
				createChannel("Channel2", "ch2", "Movies"),
			},
			selector:       &common.Selector{Type: common.SelectorName},
			patterns:       common.RegexpArr{},
			expectedGroups: map[string][]string{},
		},
		{
			name: "patterns create groups",
			channels: []*store.Channel{
				createChannel("Discovery Channel HD", "ch1", "Sports"),
				createChannel("Discovery Channel 4K", "ch2", "Movies"),
				createChannel("National Geographic UHD", "ch3", "News"),
			},
			selector: &common.Selector{Type: common.SelectorName},
			patterns: createPatterns("HD", "4K", "UHD"),
			expectedGroups: map[string][]string{
				"Discovery Channel": {"Discovery Channel HD", "Discovery Channel 4K"},
			},
		},
		{
			name: "no modified base names, no grouping",
			channels: []*store.Channel{
				createChannel("Different Channel", "ch1", "Sports"),
				createChannel("Another Channel", "ch2", "Movies"),
			},
			selector:       &common.Selector{Type: common.SelectorName},
			patterns:       createPatterns("HD"),
			expectedGroups: map[string][]string{},
		},
		{
			name: "priority sorting",
			channels: []*store.Channel{
				createChannel("CNN HD", "ch1", "News"),
				createChannel("CNN 4K", "ch2", "News"),
				createChannel("CNN", "ch3", "News"),
			},
			selector: &common.Selector{Type: common.SelectorName},
			patterns: createPatterns("", "4K", "HD"),
			expectedGroups: map[string][]string{
				"CNN": {"CNN", "CNN 4K", "CNN HD"}, // "" has highest priority, then 4K, then HD
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPatternMatcher(tt.channels, tt.selector, tt.patterns)
			groups := pm.GroupChannels()

			if len(groups) != len(tt.expectedGroups) {
				t.Errorf("Expected %d groups, got %d", len(tt.expectedGroups), len(groups))
				return
			}

			for baseName, expectedChannels := range tt.expectedGroups {
				group, exists := groups[baseName]
				if !exists {
					t.Errorf("Expected group %q not found", baseName)
					continue
				}
				if len(group) != len(expectedChannels) {
					t.Errorf("Group %q: expected %d channels, got %d", baseName, len(expectedChannels), len(group))
					continue
				}
				for i, ch := range group {
					if ch.Name() != expectedChannels[i] {
						t.Errorf("Group %q[%d]: expected %q, got %q", baseName, i, expectedChannels[i], ch.Name())
					}
				}
			}
		})
	}
}

func TestExtractBaseName(t *testing.T) {
	tests := []struct {
		name     string
		channel  *store.Channel
		selector *common.Selector
		patterns common.RegexpArr
		expected string
	}{
		{
			name:     "no patterns",
			channel:  createChannel("Channel Name", "ch1", "Group"),
			selector: &common.Selector{Type: common.SelectorName},
			patterns: common.RegexpArr{},
			expected: "Channel Name",
		},
		{
			name:     "patterns remove parts",
			channel:  createChannel("Discovery Channel HD", "ch1", "Group"),
			selector: &common.Selector{Type: common.SelectorName},
			patterns: createPatterns("HD"),
			expected: "Discovery Channel",
		},
		{
			name:     "multiple patterns",
			channel:  createChannel("[HD] Discovery Channel (FHD)", "ch1", "Group"),
			selector: &common.Selector{Type: common.SelectorName},
			patterns: createPatterns(`\[HD\]`, `\(FHD\)`),
			expected: "Discovery Channel",
		},
		{
			name: "attribute selector",
			channel: func() *store.Channel {
				track := &m3u8.Track{
					Name: "Channel",
					URI:  mustParseURL("http://example.com/stream"),
					Attrs: map[string]string{
						m3u8.AttrTvgID: "discovery.hd",
					},
					Tags: map[string]string{
						m3u8.TagGroup: "Group",
					},
				}
				return store.NewChannel(track, nil)
			}(),
			selector: &common.Selector{Type: common.SelectorAttr, Value: m3u8.AttrTvgID},
			patterns: createPatterns("hd"),
			expected: "discovery.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPatternMatcher([]*store.Channel{tt.channel}, tt.selector, tt.patterns)
			result, _ := pm.extractBaseName(tt.channel)
			if result != tt.expected {
				t.Errorf("extractBaseName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractBaseNameFromChannel(t *testing.T) {
	selector := &common.Selector{Type: common.SelectorAttr, Value: "x-tvg-name"}
	patterns := common.RegexpArr{
		regexp.MustCompile(`\[HD\]`),
		regexp.MustCompile(`\(FHD\)`),
		regexp.MustCompile(`HD`),
	}

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
			track := &m3u8.Track{
				Name: "Channel Name",
				Attrs: map[string]string{
					"x-tvg-name": tt.input,
				},
			}
			ch := store.NewChannel(track, nil)
			pm := NewPatternMatcher(nil, selector, patterns)
			result, ok := pm.extractBaseName(ch)
			if !ok {
				t.Errorf("extractBaseName(%q) failed to extract value", tt.input)
				return
			}
			if result != tt.expected {
				t.Errorf("extractBaseName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractBaseNameFromChannel_attr(t *testing.T) {
	selector := &common.Selector{
		Type:  common.SelectorAttr,
		Value: "x-tvg-name",
	}
	patterns := common.RegexpArr{
		regexp.MustCompile(`\-7 \(Pacific\)`),
		regexp.MustCompile(`\-7`),
		regexp.MustCompile(`\-5 \(Central\)`),
	}

	tests := []struct {
		name     string
		attrName string
		attrVal  string
		expected string
	}{
		{
			name:     "matches first pattern",
			attrName: "x-tvg-name",
			attrVal:  "-7 (Pacific)",
			expected: "",
		},
		{
			name:     "matches second pattern",
			attrName: "x-tvg-name",
			attrVal:  "-7",
			expected: "",
		},
		{
			name:     "matches third pattern",
			attrName: "x-tvg-name",
			attrVal:  "-5 (Central)",
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
			track := &m3u8.Track{
				Name:  "Test Channel",
				Attrs: map[string]string{tt.attrName: tt.attrVal},
			}
			ch := store.NewChannel(track, nil)
			pm := NewPatternMatcher(nil, selector, patterns)
			result, ok := pm.extractBaseName(ch)
			if !ok {
				t.Errorf("extractBaseName(attr=%q) failed to extract value", tt.attrVal)
				return
			}
			if result != tt.expected {
				t.Errorf("extractBaseName(attr=%q) = %q, want %q", tt.attrVal, result, tt.expected)
			}
		})
	}
}
