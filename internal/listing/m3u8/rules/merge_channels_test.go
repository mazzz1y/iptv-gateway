package rules

import (
	configrules "iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/parser/m3u8"
	"net/url"
	"regexp"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMergeChannelsProcessor_mergeWithSetPattern(t *testing.T) {
	rule := &configrules.MergeChannelsRule{
		NamePatterns: types.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`UHD`),
			regexp.MustCompile(`FHD`),
			regexp.MustCompile(`HD`),
		},
		SetField: mustTemplate("{{.BaseName}} Multi-Quality"),
	}

	processor := NewMergeChannelsActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")
	uri3, _ := url.Parse("http://example.com/url3")
	uri4, _ := url.Parse("http://example.com/url4")

	ch1 := &Channel{track: &m3u8.Track{Name: "CNN HD", URI: uri1}}
	ch2 := &Channel{track: &m3u8.Track{Name: "CNN 4K", URI: uri2}}
	ch3 := &Channel{track: &m3u8.Track{Name: "ESPN UHD", URI: uri3}}
	ch4 := &Channel{track: &m3u8.Track{Name: "Fox News", URI: uri4}}

	store.Add(ch1)
	store.Add(ch2)
	store.Add(ch3)
	store.Add(ch4)

	processor.Apply(store)

	channels := store.All()
	expectedNames := map[string]bool{
		"CNN Multi-Quality":  false,
		"ESPN Multi-Quality": false,
		"Fox News":           false,
	}

	actualNames := make([]string, 0)
	for _, ch := range channels {
		actualNames = append(actualNames, ch.Name())
		expectedNames[ch.Name()] = true
	}

	if actualNames[0] != "CNN Multi-Quality" {
		t.Errorf("Expected 'CNN Multi-Quality', got '%s'", actualNames[0])
	}
	if actualNames[1] != "CNN Multi-Quality" {
		t.Errorf("Expected 'CNN Multi-Quality', got '%s'", actualNames[1])
	}
	if actualNames[2] != "ESPN UHD" {
		t.Errorf("Expected 'ESPN UHD', got '%s'", actualNames[2])
	}
	if actualNames[3] != "Fox News" {
		t.Errorf("Expected 'Fox News', got '%s'", actualNames[3])
	}
}

func TestMergeChannelsProcessor_mergeWithoutSetPattern(t *testing.T) {
	rule := &configrules.MergeChannelsRule{
		NamePatterns: types.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`HD`),
		},
	}

	processor := NewMergeChannelsActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	ch1 := &Channel{track: &m3u8.Track{Name: "CNN HD", URI: uri1}}
	ch2 := &Channel{track: &m3u8.Track{Name: "CNN 4K", URI: uri2}}

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	channels := store.All()
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
		return
	}

	if channels[0].Name() != "CNN 4K" {
		t.Errorf("Expected 'CNN 4K', got '%s'", channels[0].Name())
	}
	if channels[1].Name() != "CNN 4K" {
		t.Errorf("Expected 'CNN 4K', got '%s'", channels[1].Name())
	}
}

func TestMergeChannelsProcessor_mergeByAttr(t *testing.T) {
	rule := &configrules.MergeChannelsRule{
		AttrPatterns: &types.NamePatterns{
			Name: "tvg-id",
			Patterns: types.RegexpArr{
				regexp.MustCompile(`HD`),
				regexp.MustCompile(`4K`),
			},
		},
		SetField: mustTemplate("{{.BaseName}} Merged"),
	}

	processor := NewMergeChannelsActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	ch1 := &Channel{track: &m3u8.Track{
		Name:  "Channel 1",
		URI:   uri1,
		Attrs: map[string]string{"tvg-id": "Test HD"},
	}}
	ch2 := &Channel{track: &m3u8.Track{
		Name:  "Channel 2",
		URI:   uri2,
		Attrs: map[string]string{"tvg-id": "Test 4K"},
	}}

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	channels := store.All()
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
		return
	}

	attr0, _ := channels[0].GetAttr("tvg-id")
	if attr0 != "Test Merged" {
		t.Errorf("Expected 'Test Merged', got '%s'", attr0)
	}
	attr1, _ := channels[1].GetAttr("tvg-id")
	if attr1 != "Test Merged" {
		t.Errorf("Expected 'Test Merged', got '%s'", attr1)
	}
}

func TestMergeChannelsProcessor_noPatternMatch(t *testing.T) {
	rule := &configrules.MergeChannelsRule{
		NamePatterns: types.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`HD`),
		},
		SetField: mustTemplate("Merged"),
	}

	processor := NewMergeChannelsActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")

	ch1 := &Channel{track: &m3u8.Track{Name: "CNN Standard", URI: uri1}}

	store.Add(ch1)

	processor.Apply(store)

	channels := store.All()
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
		return
	}

	if channels[0].Name() != "CNN Standard" {
		t.Errorf("Expected 'CNN Standard', got '%s'", channels[0].Name())
	}
}

func mustTemplate(tmpl string) *types.Template {
	var t types.Template
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: tmpl,
	}
	if err := t.UnmarshalYAML(node); err != nil {
		panic(err)
	}
	return &t
}
