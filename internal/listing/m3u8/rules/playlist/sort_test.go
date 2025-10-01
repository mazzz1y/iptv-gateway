package playlist

import (
	"iptv-gateway/internal/config/common"
	configrules "iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/listing/m3u8/store"
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
	channels := []*store.Channel{
		store.NewChannel(&m3u8.Track{Name: "ZZZ Channel"}, nil),
		store.NewChannel(&m3u8.Track{Name: "AAA Channel"}, nil),
		store.NewChannel(&m3u8.Track{Name: "MMM Channel"}, nil),
		store.NewChannel(&m3u8.Track{Name: "BBB Channel"}, nil),
	}

	s := store.NewStore()
	for _, ch := range channels {
		s.Add(ch)
	}

	rule := &configrules.Sort{}
	processor := NewSortProcessor(rule)

	processor.Apply(s)

	sorted := s.All()

	expected := []string{"AAA Channel", "BBB Channel", "MMM Channel", "ZZZ Channel"}
	for i, ch := range sorted {
		if ch.Name() != expected[i] {
			t.Errorf("Expected channel %d to be %q, got %q", i, expected[i], ch.Name())
		}
	}
}

func TestSortProcessor_Apply_WithOrder(t *testing.T) {
	channels := []*store.Channel{
		store.NewChannel(&m3u8.Track{Name: "News Channel"}, nil),
		store.NewChannel(&m3u8.Track{Name: "Sports Channel"}, nil),
		store.NewChannel(&m3u8.Track{Name: "Music Channel"}, nil),
		store.NewChannel(&m3u8.Track{Name: "Movie Channel"}, nil),
	}

	s := store.NewStore()
	for _, ch := range channels {
		s.Add(ch)
	}

	order := mustCompileRegexpArr([]string{"Sports.*", "Music.*", ""})
	rule := &configrules.Sort{
		Order: &order,
	}
	processor := NewSortProcessor(rule)

	processor.Apply(s)

	sorted := s.All()

	expected := []string{"Sports Channel", "Music Channel", "Movie Channel", "News Channel"}
	for i, ch := range sorted {
		if ch.Name() != expected[i] {
			t.Errorf("Expected channel %d to be %q, got %q", i, expected[i], ch.Name())
		}
	}
}

func TestSortProcessor_Apply_WithGroupBy(t *testing.T) {
	channels := []*store.Channel{
		store.NewChannel(&m3u8.Track{
			Name:  "Sports 1",
			Attrs: map[string]string{"group-title": "Sports"},
		}, nil),
		store.NewChannel(&m3u8.Track{
			Name:  "News 1",
			Attrs: map[string]string{"group-title": "News"},
		}, nil),
		store.NewChannel(&m3u8.Track{
			Name:  "Sports 2",
			Attrs: map[string]string{"group-title": "Sports"},
		}, nil),
		store.NewChannel(&m3u8.Track{
			Name:  "Music 1",
			Attrs: map[string]string{"group-title": "Music"},
		}, nil),
	}

	s := store.NewStore()
	for _, ch := range channels {
		s.Add(ch)
	}

	groupOrder := mustCompileRegexpArr([]string{"News", "Sports", "Music"})
	rule := &configrules.Sort{
		GroupBy: &configrules.GroupByRule{
			Selector: &common.Selector{Type: common.SelectorAttr, Value: "group-title"},
			Order:    &groupOrder,
		},
	}
	processor := NewSortProcessor(rule)

	processor.Apply(s)

	sorted := s.All()

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
