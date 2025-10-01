package playlist

import (
	"iptv-gateway/internal/config/common"
	configrules "iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/listing/m3u8/store"
	"iptv-gateway/internal/parser/m3u8"
	"net/url"
	"regexp"
	"testing"
)

type mockDuplicatesMatcher struct {
	groups map[string][]*store.Channel
}

func (m *mockDuplicatesMatcher) GroupChannels() map[string][]*store.Channel {
	return m.groups
}

func TestRemoveDuplicatesProcessor_shouldNotRemoveIdenticalChannels(t *testing.T) {
	rule := &configrules.RemoveDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName}, Patterns: common.RegexpArr{
			regexp.MustCompile(`\[HD\]`),
			regexp.MustCompile(`\(FHD\)`),
		},
	}

	processor := NewRemoveDuplicatesActionProcessor(rule)
	s := store.NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")
	uri3, _ := url.Parse("http://example.com/url3")
	uri4, _ := url.Parse("http://example.com/url4")

	ch1 := store.NewChannel(&m3u8.Track{Name: "Channel Name [HD]", URI: uri1}, nil)
	ch2 := store.NewChannel(&m3u8.Track{Name: "Channel Name (FHD)", URI: uri2}, nil)

	ch3 := store.NewChannel(&m3u8.Track{Name: "Different Channel", URI: uri3}, nil)
	ch4 := store.NewChannel(&m3u8.Track{Name: "Different Channel", URI: uri4}, nil)

	s.Add(ch1)
	s.Add(ch2)
	s.Add(ch3)
	s.Add(ch4)

	processor.Apply(s)

	activeChannels := make([]*store.Channel, 0)
	for _, ch := range s.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	expectedActive := 3
	if len(activeChannels) != expectedActive {
		t.Errorf("Expected %d active channels, got %d", expectedActive, len(activeChannels))

		for i, ch := range s.All() {
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

	s := store.NewStore()

	playlist := mockPlaylist{name: "test-playlist"}

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")
	uri3, _ := url.Parse("http://example.com/url3")
	uri4, _ := url.Parse("http://example.com/url4")

	track1 := &m3u8.Track{Name: "Discovery Channel HD", URI: uri1}
	track2 := &m3u8.Track{Name: "Discovery Channel 4K", URI: uri2}
	track3 := &m3u8.Track{Name: "National Geographic UHD", URI: uri3}
	track4 := &m3u8.Track{Name: "National Geographic", URI: uri4}

	ch1 := store.NewChannel(track1, playlist)
	ch2 := store.NewChannel(track2, playlist)
	ch3 := store.NewChannel(track3, playlist)
	ch4 := store.NewChannel(track4, playlist)

	s.Add(ch1)
	s.Add(ch2)
	s.Add(ch3)
	s.Add(ch4)

	processor := &RemoveDuplicatesProcessor{
		rule: rule,
		matcher: &mockDuplicatesMatcher{
			groups: map[string][]*store.Channel{
				"Discovery Channel":   {ch2, ch1}, // ch2 4K best
				"National Geographic": {ch4, ch3}, // ch4 no pattern best
			},
		},
	}

	processor.Apply(s)

	activeChannels := make([]*store.Channel, 0)
	for _, ch := range s.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	expectedActive := 2
	if len(activeChannels) != expectedActive {
		t.Errorf("Expected %d active channels, got %d", expectedActive, len(activeChannels))
		for i, ch := range s.All() {
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
	s := store.NewStore()

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

	ch1 := store.NewChannel(track1, playlist)
	ch2 := store.NewChannel(track2, playlist)

	s.Add(ch1)
	s.Add(ch2)

	processor.Apply(s)

	activeChannels := make([]*store.Channel, 0)
	for _, ch := range s.All() {
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
	s := store.NewStore()

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

	ch1 := store.NewChannel(track1, playlist)
	ch2 := store.NewChannel(track2, playlist)

	s.Add(ch1)
	s.Add(ch2)

	processor.Apply(s)

	activeChannels := make([]*store.Channel, 0)
	for _, ch := range s.All() {
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
	s := store.NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")
	uri3, _ := url.Parse("http://example.com/url3")
	uri4, _ := url.Parse("http://example.com/url4")
	uri5, _ := url.Parse("http://example.com/url5")
	uri6, _ := url.Parse("http://example.com/url6")

	ch1 := store.NewChannel(&m3u8.Track{Name: "Channel A HD", URI: uri1}, nil)
	ch2 := store.NewChannel(&m3u8.Track{Name: "Channel A 4K", URI: uri2}, nil)
	ch3 := store.NewChannel(&m3u8.Track{Name: "Channel B", URI: uri3}, nil)
	ch4 := store.NewChannel(&m3u8.Track{Name: "Channel B", URI: uri4}, nil)
	ch5 := store.NewChannel(&m3u8.Track{Name: "Channel C orig", URI: uri5}, nil)
	ch6 := store.NewChannel(&m3u8.Track{Name: "Channel C", URI: uri6}, nil)

	s.Add(ch1)
	s.Add(ch2)
	s.Add(ch3)
	s.Add(ch4)
	s.Add(ch5)
	s.Add(ch6)

	processor.Apply(s)

	activeChannels := make([]*store.Channel, 0)
	for _, ch := range s.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	expectedActive := 4
	if len(activeChannels) != expectedActive {
		t.Errorf("Expected %d active channels, got %d", expectedActive, len(activeChannels))
		for i, ch := range s.All() {
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
	s := store.NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	ch1 := store.NewChannel(&m3u8.Track{Name: "Test Channel", URI: uri1}, nil)
	ch2 := store.NewChannel(&m3u8.Track{Name: "Test Channel orig", URI: uri2}, nil)

	s.Add(ch1)
	s.Add(ch2)

	processor.Apply(s)

	activeChannels := make([]*store.Channel, 0)
	for _, ch := range s.All() {
		if !ch.IsRemoved() {
			activeChannels = append(activeChannels, ch)
		}
	}

	expectedActive := 1
	if len(activeChannels) != expectedActive {
		t.Errorf("Expected %d active channels, got %d", expectedActive, len(activeChannels))
		for i, ch := range s.All() {
			t.Logf("Channel %d: Name='%s', URI='%s', Removed=%v",
				i, ch.Name(), ch.URI(), ch.IsRemoved())
		}
		return
	}

	if activeChannels[0].Name() != "Test Channel" {
		t.Errorf("Expected 'Test Channel' to be kept (empty pattern has higher priority than 'orig'), got '%s'", activeChannels[0].Name())
	}
}
