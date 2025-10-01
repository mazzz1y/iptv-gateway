package rules

import (
	"iptv-gateway/internal/config/common"
	configrules "iptv-gateway/internal/config/rules/playlist"
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
		Selector: &common.Selector{Type: common.SelectorAttr, Value: "x-tvg-name"},
		Patterns: common.RegexpArr{
			regexp.MustCompile(`\[HD\]`),
			regexp.MustCompile(`\(FHD\)`),
			regexp.MustCompile(`HD`),
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
					Name: "Channel Name",
					Attrs: map[string]string{
						"x-tvg-name": tt.input,
					},
				},
			}
			result, ok := extractBaseNameFromChannel(ch, processor.rule.Selector, processor.rule.Patterns)
			if !ok {
				t.Errorf("extractBaseNameFromChannel(%q) failed to extract value", tt.input)
				return
			}
			if result != tt.expected {
				t.Errorf("extractKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveDuplicatesProcessor_extractKey_attr(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		Selector: &common.Selector{
			Type:  common.SelectorAttr,
			Value: "x-tvg-name",
		},
		Patterns: common.RegexpArr{
			regexp.MustCompile(`\+3 \(Омск\)`),
			regexp.MustCompile(`\+3`),
			regexp.MustCompile(`\+7 \(Москва\)`),
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
			result, ok := extractBaseNameFromChannel(ch, processor.rule.Selector, processor.rule.Patterns)
			if !ok {
				t.Errorf("extractBaseNameFromChannel(attr=%q) failed to extract value", tt.attrVal)
				return
			}
			if result != tt.expected {
				t.Errorf("extractKey(attr=%q) = %q, want %q", tt.attrVal, result, tt.expected)
			}
		})
	}
}

func TestRemoveDuplicatesProcessor_shouldNotRemoveIdenticalChannels(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName}, Patterns: common.RegexpArr{
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
		Selector: &common.Selector{Type: common.SelectorName}, Patterns: common.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`UHD`),
			regexp.MustCompile(`FHD`),
			regexp.MustCompile(`HD`),
			regexp.MustCompile(``),
		},
		FinalValue: &configrules.RemoveDuplicatesFinalValue{
			Selector: &common.Selector{Type: common.SelectorName},
			Template: mustTemplate("{{.Channel.BaseName}} HQ-Preferred"),
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)
	store := NewStore()

	playlist := mockPlaylist{name: "test-playlist"}

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")
	uri3, _ := url.Parse("http://example.com/url3")
	uri4, _ := url.Parse("http://example.com/url4")

	track1 := &m3u8.Track{Name: "Discovery Channel HD", URI: uri1}
	track2 := &m3u8.Track{Name: "Discovery Channel 4K", URI: uri2}
	track3 := &m3u8.Track{Name: "National Geographic UHD", URI: uri3}
	track4 := &m3u8.Track{Name: "National Geographic", URI: uri4}

	ch1 := NewChannel(track1, playlist)
	ch2 := NewChannel(track2, playlist)
	ch3 := NewChannel(track3, playlist)
	ch4 := NewChannel(track4, playlist)

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

func TestRemoveDuplicatesProcessor_setFieldAttr(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName}, Patterns: common.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`HD`),
		},
		FinalValue: &configrules.RemoveDuplicatesFinalValue{
			Selector: &common.Selector{
				Type:  common.SelectorAttr,
				Value: "group-title",
			},
			Template: mustTemplate("{{.Channel.BaseName}} Group"),
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)
	store := NewStore()

	playlist := mockPlaylist{name: "test-playlist"}

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	track1 := &m3u8.Track{
		Name:  "CNN HD",
		URI:   uri1,
		Attrs: map[string]string{"group-title": "News HD"},
	}
	track2 := &m3u8.Track{
		Name:  "CNN 4K",
		URI:   uri2,
		Attrs: map[string]string{"group-title": "News 4K"},
	}

	ch1 := NewChannel(track1, playlist)
	ch2 := NewChannel(track2, playlist)

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	activeChannels := make([]*Channel, 0)
	for _, ch := range store.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	if len(activeChannels) != 1 {
		t.Errorf("Expected 1 active channel, got %d", len(activeChannels))
		return
	}

	ch := activeChannels[0]
	if ch.Name() != "CNN 4K" {
		t.Errorf("Expected 'CNN 4K' as the best channel, got '%s'", ch.Name())
	}

	groupTitle, exists := ch.GetAttr("group-title")
	if !exists {
		t.Error("Expected group-title attribute to exist")
		return
	}
	if groupTitle != "CNN Group" {
		t.Errorf("Expected group-title 'CNN Group', got '%s'", groupTitle)
	}
}

func TestRemoveDuplicatesProcessor_setFieldTag(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName}, Patterns: common.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`HD`),
		},
		FinalValue: &configrules.RemoveDuplicatesFinalValue{
			Selector: &common.Selector{
				Type:  common.SelectorTag,
				Value: "quality",
			},
			Template: mustTemplate("{{.Channel.BaseName}} Multi"),
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)
	store := NewStore()

	playlist := mockPlaylist{name: "test-playlist"}

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	track1 := &m3u8.Track{
		Name:  "CNN HD",
		URI:   uri1,
		Attrs: map[string]string{"group-title": "News HD"},
		Tags:  map[string]string{"quality": "HD"},
	}
	track2 := &m3u8.Track{
		Name:  "CNN 4K",
		URI:   uri2,
		Attrs: map[string]string{"group-title": "News 4K"},
		Tags:  map[string]string{"quality": "4K"},
	}

	ch1 := NewChannel(track1, playlist)
	ch2 := NewChannel(track2, playlist)

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	activeChannels := make([]*Channel, 0)
	for _, ch := range store.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	if len(activeChannels) != 1 {
		t.Errorf("Expected 1 active channel, got %d", len(activeChannels))
		return
	}

	ch := activeChannels[0]
	if ch.Name() != "CNN 4K" {
		t.Errorf("Expected 'CNN 4K' as the best channel, got '%s'", ch.Name())
	}

	quality, exists := ch.GetTag("quality")
	if !exists {
		t.Error("Expected quality tag to exist")
		return
	}
	if quality != "CNN Multi" {
		t.Errorf("Expected quality tag 'CNN Multi', got '%s'", quality)
	}
}

func TestRemoveDuplicatesProcessor_onlyPatternChannels(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName}, Patterns: common.RegexpArr{
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
		Selector: &common.Selector{Type: common.SelectorName}, Patterns: common.RegexpArr{
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
