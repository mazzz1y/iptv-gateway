package rules

import (
	configrules "iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/parser/m3u8"
	"net/url"
	"regexp"
	"testing"
)

func stringPtr(s string) *string {
	return &s
}

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

func TestRemoveDuplicatesProcessor_shouldNotRemoveIdenticalChannels(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		NamePatterns: types.RegexpArr{
			regexp.MustCompile(`\[HD\]`),
			regexp.MustCompile(`\(FHD\)`),
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")
	uri3, _ := url.Parse("http://example.com/url3")
	uri4, _ := url.Parse("http://example.com/url4")

	ch1 := &Channel{track: &m3u8.Track{Name: "Channel Name [HD]", URI: uri1}}
	ch2 := &Channel{track: &m3u8.Track{Name: "Channel Name (FHD)", URI: uri2}}

	ch3 := &Channel{track: &m3u8.Track{Name: "Different Channel", URI: uri3}}
	ch4 := &Channel{track: &m3u8.Track{Name: "Different Channel", URI: uri4}}

	store.Add(ch1)
	store.Add(ch2)
	store.Add(ch3)
	store.Add(ch4)

	processor.Apply(store)

	activeChannels := make([]*Channel, 0)
	for _, ch := range store.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	expectedActive := 3
	if len(activeChannels) != expectedActive {
		t.Errorf("Expected %d active channels, got %d", expectedActive, len(activeChannels))

		for i, ch := range store.All() {
			t.Logf("Channel %d: Name='%s', URI='%s', Removed=%v",
				i, ch.Name(), ch.URI(), ch.IsRemoved())
		}
	}

	identicalChannelsActive := 0
	for _, ch := range activeChannels {
		if ch.Name() == "Different Channel" {
			identicalChannelsActive++
		}
	}

	if identicalChannelsActive != 2 {
		t.Errorf("Expected 2 identical channels to remain active, got %d", identicalChannelsActive)
	}
}

func TestRemoveDuplicatesProcessor_setPattern(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		NamePatterns: types.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`UHD`),
			regexp.MustCompile(`FHD`),
			regexp.MustCompile(`HD`),
			regexp.MustCompile(``),
		},
		SetField: mustTemplate("{{.BaseName}} HQ-Preferred"),
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")
	uri3, _ := url.Parse("http://example.com/url3")
	uri4, _ := url.Parse("http://example.com/url4")

	ch1 := &Channel{track: &m3u8.Track{Name: "Discovery Channel HD", URI: uri1}}
	ch2 := &Channel{track: &m3u8.Track{Name: "Discovery Channel 4K", URI: uri2}}
	ch3 := &Channel{track: &m3u8.Track{Name: "National Geographic UHD", URI: uri3}}
	ch4 := &Channel{track: &m3u8.Track{Name: "National Geographic", URI: uri4}}

	store.Add(ch1)
	store.Add(ch2)
	store.Add(ch3)
	store.Add(ch4)

	processor.Apply(store)

	activeChannels := make([]*Channel, 0)
	for _, ch := range store.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	expectedActive := 2
	if len(activeChannels) != expectedActive {
		t.Errorf("Expected %d active channels, got %d", expectedActive, len(activeChannels))
		for i, ch := range store.All() {
			t.Logf("Channel %d: Name='%s', URI='%s', Removed=%v",
				i, ch.Name(), ch.URI(), ch.IsRemoved())
		}
	}

	expectedNames := map[string]bool{
		"Discovery Channel HQ-Preferred":   false,
		"National Geographic HQ-Preferred": false,
	}

	for _, ch := range activeChannels {
		if _, exists := expectedNames[ch.Name()]; exists {
			expectedNames[ch.Name()] = true
		} else {
			t.Errorf("Unexpected channel name: %s", ch.Name())
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected channel name not found: %s", name)
		}
	}
}

func TestRemoveDuplicatesProcessor_onlyPatternChannels(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		NamePatterns: types.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`UHD`),
			regexp.MustCompile(`HD 50`),
			regexp.MustCompile(`HD 50 orig`),
			regexp.MustCompile(`FHD`),
			regexp.MustCompile(`HD`),
			regexp.MustCompile(`HD orig`),
			regexp.MustCompile(`orig`),
			regexp.MustCompile(``),
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")
	uri3, _ := url.Parse("http://example.com/url3")
	uri4, _ := url.Parse("http://example.com/url4")
	uri5, _ := url.Parse("http://example.com/url5")
	uri6, _ := url.Parse("http://example.com/url6")

	ch1 := &Channel{track: &m3u8.Track{Name: "Channel A HD", URI: uri1}}
	ch2 := &Channel{track: &m3u8.Track{Name: "Channel A 4K", URI: uri2}}
	ch3 := &Channel{track: &m3u8.Track{Name: "Channel B", URI: uri3}}
	ch4 := &Channel{track: &m3u8.Track{Name: "Channel B", URI: uri4}}
	ch5 := &Channel{track: &m3u8.Track{Name: "Channel C orig", URI: uri5}}
	ch6 := &Channel{track: &m3u8.Track{Name: "Channel C", URI: uri6}}

	store.Add(ch1)
	store.Add(ch2)
	store.Add(ch3)
	store.Add(ch4)
	store.Add(ch5)
	store.Add(ch6)

	processor.Apply(store)

	activeChannels := make([]*Channel, 0)
	for _, ch := range store.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	expectedActive := 4
	if len(activeChannels) != expectedActive {
		t.Errorf("Expected %d active channels, got %d", expectedActive, len(activeChannels))
		for i, ch := range store.All() {
			t.Logf("Channel %d: Name='%s', URI='%s', Removed=%v",
				i, ch.Name(), ch.URI(), ch.IsRemoved())
		}
		return
	}

	channelBCount := 0
	for _, ch := range activeChannels {
		if ch.Name() == "Channel B" {
			channelBCount++
		}
	}

	if channelBCount != 2 {
		t.Errorf("Expected 2 'Channel B' channels (no patterns, should not be deduplicated), got %d", channelBCount)
	}

	hasChannelA4K := false
	hasChannelC := false
	for _, ch := range activeChannels {
		if ch.Name() == "Channel A 4K" {
			hasChannelA4K = true
		}
		if ch.Name() == "Channel C" {
			hasChannelC = true
		}
	}

	if !hasChannelA4K {
		t.Error("Expected 'Channel A 4K' to be kept (higher priority than HD)")
	}

	hasChannelCOrig := false
	for _, ch := range activeChannels {
		if ch.Name() == "Channel C orig" {
			hasChannelCOrig = true
		}
	}

	if hasChannelC {
		t.Error("Expected 'Channel C' to be removed")
	}
	if !hasChannelCOrig {
		t.Error("Expected 'Channel C orig' to be kept")
	}
}

func TestRemoveDuplicatesProcessor_emptyPatternPriority(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		NamePatterns: types.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`UHD`),
			regexp.MustCompile(`HD 50`),
			regexp.MustCompile(`HD 50 orig`),
			regexp.MustCompile(`FHD`),
			regexp.MustCompile(`HD`),
			regexp.MustCompile(`HD orig`),
			regexp.MustCompile(``),
			regexp.MustCompile(`orig`),
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	ch1 := &Channel{track: &m3u8.Track{Name: "Test Channel", URI: uri1}}
	ch2 := &Channel{track: &m3u8.Track{Name: "Test Channel orig", URI: uri2}}

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	activeChannels := make([]*Channel, 0)
	for _, ch := range store.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	expectedActive := 1
	if len(activeChannels) != expectedActive {
		t.Errorf("Expected %d active channels, got %d", expectedActive, len(activeChannels))
		for i, ch := range store.All() {
			t.Logf("Channel %d: Name='%s', URI='%s', Removed=%v",
				i, ch.Name(), ch.URI(), ch.IsRemoved())
		}
		return
	}

	if activeChannels[0].Name() != "Test Channel" {
		t.Errorf("Expected 'Test Channel' to be kept (empty pattern has higher priority than 'orig'), got '%s'", activeChannels[0].Name())
	}
}
