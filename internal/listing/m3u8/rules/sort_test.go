package rules

import (
	configrules "iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/parser/m3u8"
	"testing"
)

func TestSortProcessor_Apply_SimpleSort(t *testing.T) {
	channels := []*Channel{
		{track: &m3u8.Track{Name: "ZZZ Channel"}},
		{track: &m3u8.Track{Name: "AAA Channel"}},
		{track: &m3u8.Track{Name: "MMM Channel"}},
		{track: &m3u8.Track{Name: "BBB Channel"}},
	}

	store := NewStore()
	for _, ch := range channels {
		store.Add(ch)
	}

	rule := &configrules.SortRule{}
	processor := NewSortProcessor(rule)

	processor.Apply(store)

	sorted := store.All()

	expected := []string{"AAA Channel", "BBB Channel", "MMM Channel", "ZZZ Channel"}
	for i, ch := range sorted {
		if ch.Name() != expected[i] {
			t.Errorf("Expected channel %d to be %q, got %q", i, expected[i], ch.Name())
		}
	}
}

func TestSortProcessor_Apply_WithOrder(t *testing.T) {
	channels := []*Channel{
		{track: &m3u8.Track{Name: "News Channel"}},
		{track: &m3u8.Track{Name: "Sports Channel"}},
		{track: &m3u8.Track{Name: "Music Channel"}},
		{track: &m3u8.Track{Name: "Movie Channel"}},
	}

	store := NewStore()
	for _, ch := range channels {
		store.Add(ch)
	}

	order := types.StringOrArr{"Sports.*", "Music.*", ""}
	rule := &configrules.SortRule{
		Order: &order,
	}
	processor := NewSortProcessor(rule)

	processor.Apply(store)

	sorted := store.All()

	expected := []string{"Sports Channel", "Music Channel", "Movie Channel", "News Channel"}
	for i, ch := range sorted {
		if ch.Name() != expected[i] {
			t.Errorf("Expected channel %d to be %q, got %q", i, expected[i], ch.Name())
		}
	}
}

func TestSortProcessor_Apply_WithGroupBy(t *testing.T) {
	channels := []*Channel{
		{track: &m3u8.Track{
			Name:  "Sports 1",
			Attrs: map[string]string{"group-title": "Sports"},
		}},
		{track: &m3u8.Track{
			Name:  "News 1",
			Attrs: map[string]string{"group-title": "News"},
		}},
		{track: &m3u8.Track{
			Name:  "Sports 2",
			Attrs: map[string]string{"group-title": "Sports"},
		}},
		{track: &m3u8.Track{
			Name:  "Music 1",
			Attrs: map[string]string{"group-title": "Music"},
		}},
	}

	store := NewStore()
	for _, ch := range channels {
		store.Add(ch)
	}

	groupOrder := types.StringOrArr{"News", "Sports", "Music"}
	rule := &configrules.SortRule{
		GroupBy: &configrules.GroupByRule{
			Attr:  "group-title",
			Order: &groupOrder,
		},
	}
	processor := NewSortProcessor(rule)

	processor.Apply(store)

	sorted := store.All()

	expectedGroups := []string{"News", "Sports", "Sports", "Music"}
	for i, ch := range sorted {
		groupValue, _ := ch.GetAttr("group-title")
		if groupValue != expectedGroups[i] {
			t.Errorf("Expected channel %d to be in group %q, got %q", i, expectedGroups[i], groupValue)
		}
	}

	expectedNames := []string{"News 1", "Sports 1", "Sports 2", "Music 1"}
	for i, ch := range sorted {
		if ch.Name() != expectedNames[i] {
			t.Errorf("Expected channel %d to be %q, got %q", i, expectedNames[i], ch.Name())
		}
	}
}

func TestSortProcessor_getChannelSortValue(t *testing.T) {
	tests := []struct {
		name     string
		rule     *configrules.SortRule
		channel  *Channel
		expected string
	}{
		{
			name: "no attr or tag specified",
			rule: &configrules.SortRule{},
			channel: &Channel{
				track: &m3u8.Track{Name: "Test Channel"},
			},
			expected: "Test Channel",
		},
		{
			name: "with attr specified",
			rule: &configrules.SortRule{
				Attr: "tvg-name",
			},
			channel: &Channel{
				track: &m3u8.Track{
					Name:  "Test Channel",
					Attrs: map[string]string{"tvg-name": "Custom Name"},
				},
			},
			expected: "Custom Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewSortProcessor(tt.rule)
			result := processor.getChannelSortValue(tt.channel)
			if result != tt.expected {
				t.Errorf("getChannelSortValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSortProcessor_matchesPattern(t *testing.T) {
	processor := NewSortProcessor(&configrules.SortRule{})

	tests := []struct {
		name     string
		value    string
		pattern  string
		expected bool
	}{
		{
			name:     "empty pattern matches anything",
			value:    "any value",
			pattern:  "",
			expected: true,
		},
		{
			name:     "exact match",
			value:    "Sports Channel",
			pattern:  "Sports Channel",
			expected: true,
		},
		{
			name:     "regex match",
			value:    "Sports Channel",
			pattern:  "Sports.*",
			expected: true,
		},
		{
			name:     "no match",
			value:    "News Channel",
			pattern:  "Sports.*",
			expected: false,
		},
		{
			name:     "invalid regex falls back to exact match",
			value:    "test[",
			pattern:  "test[",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.matchesPattern(tt.value, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.value, tt.pattern, result, tt.expected)
			}
		})
	}
}
