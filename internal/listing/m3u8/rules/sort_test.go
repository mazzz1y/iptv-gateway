package rules

import (
	"iptv-gateway/internal/config/common"
	configrules "iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/parser/m3u8"
	"regexp"
	"testing"
)

func mustCompileRegexpArr(patterns []string) common.RegexpArr {
	result := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		if pattern == "" {
			result[i] = nil
		} else {
			result[i] = regexp.MustCompile(pattern)
		}
	}
	return result
}

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

	rule := &configrules.Sort{}
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

	order := mustCompileRegexpArr([]string{"Sports.*", "Music.*", ""})
	rule := &configrules.Sort{
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

	groupOrder := mustCompileRegexpArr([]string{"News", "Sports", "Music"})
	rule := &configrules.Sort{
		GroupBy: &configrules.GroupByRule{
			Selector: &common.Selector{Type: common.SelectorAttr, Value: "group-title"},
			Order:    &groupOrder,
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
