package rules

import (
	"iptv-gateway/internal/config/common"
	configrules "iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/parser/m3u8"
	"net/url"
	"regexp"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMergeChannelsProcessor_CopyTvgId(t *testing.T) {
	rule := &configrules.MergeDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName},
		Patterns: common.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`HD`),
		},
	}

	processor := NewMergeDuplicatesActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	ch1 := &Channel{track: &m3u8.Track{
		Name:  "CNN HD",
		URI:   uri1,
		Attrs: map[string]string{"tvg-id": "cnn-hd"},
	}}
	ch2 := &Channel{track: &m3u8.Track{
		Name:  "CNN 4K",
		URI:   uri2,
		Attrs: map[string]string{"tvg-id": "cnn-4k"},
	}}

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	channels := store.All()
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
		return
	}

	for _, ch := range channels {
		tvgID, exists := ch.GetAttr("tvg-id")
		if !exists {
			t.Errorf("Expected tvg-id attribute to exist")
			continue
		}
		if tvgID != "cnn-4k" {
			t.Errorf("Expected tvg-id 'cnn-4k', got '%s'", tvgID)
		}
	}
}

func TestMergeChannelsProcessor_SetFieldName(t *testing.T) {
	rule := &configrules.MergeDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName},
		Patterns: common.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`HD`),
		},
		FinalValue: &configrules.MergeDuplicatesFinalValue{
			Selector: &common.Selector{Type: common.SelectorName},
			Template: mustTemplate("{{.BaseName}} Multi-Quality"),
		},
	}

	processor := NewMergeDuplicatesActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	ch1 := &Channel{track: &m3u8.Track{
		Name:  "CNN HD",
		URI:   uri1,
		Attrs: map[string]string{"tvg-id": "cnn-hd"},
	}}
	ch2 := &Channel{track: &m3u8.Track{
		Name:  "CNN 4K",
		URI:   uri2,
		Attrs: map[string]string{"tvg-id": "cnn-4k"},
	}}

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	channels := store.All()
	for _, ch := range channels {
		if ch.Name() != "CNN Multi-Quality" {
			t.Errorf("Expected name 'CNN Multi-Quality', got '%s'", ch.Name())
		}
	}
}

func TestMergeChannelsProcessor_SetFieldAttr(t *testing.T) {
	rule := &configrules.MergeDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName},
		Patterns: common.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`HD`),
		},
		FinalValue: &configrules.MergeDuplicatesFinalValue{
			Selector: &common.Selector{Type: common.SelectorAttr, Value: "group-title"},
			Template: mustTemplate("{{.BaseName}} Group"),
		},
	}

	processor := NewMergeDuplicatesActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	ch1 := &Channel{track: &m3u8.Track{
		Name:  "CNN HD",
		URI:   uri1,
		Attrs: map[string]string{"tvg-id": "cnn-hd", "group-title": "News HD"},
	}}
	ch2 := &Channel{track: &m3u8.Track{
		Name:  "CNN 4K",
		URI:   uri2,
		Attrs: map[string]string{"tvg-id": "cnn-4k", "group-title": "News 4K"},
	}}

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	channels := store.All()
	for _, ch := range channels {
		groupTitle, exists := ch.GetAttr("group-title")
		if !exists {
			t.Errorf("Expected group-title attribute to exist")
			continue
		}
		if groupTitle != "CNN Group" {
			t.Errorf("Expected group-title 'CNN Group', got '%s'", groupTitle)
		}

		tvgID, exists := ch.GetAttr("tvg-id")
		if !exists {
			t.Errorf("Expected tvg-id attribute to exist")
			continue
		}
		if tvgID != "cnn-4k" {
			t.Errorf("Expected tvg-id 'cnn-4k', got '%s'", tvgID)
		}
	}
}

func TestMergeChannelsProcessor_SetFieldTag(t *testing.T) {
	rule := &configrules.MergeDuplicatesRule{
		Selector: &common.Selector{Type: common.SelectorName},
		Patterns: common.RegexpArr{
			regexp.MustCompile(`4K`),
			regexp.MustCompile(`HD`),
		},
		FinalValue: &configrules.MergeDuplicatesFinalValue{
			Selector: &common.Selector{Type: common.SelectorTag, Value: "quality"},
			Template: mustTemplate("{{.BaseName}} Multi"),
		},
	}

	processor := NewMergeDuplicatesActionProcessor(rule)
	store := NewStore()

	uri1, _ := url.Parse("http://example.com/url1")
	uri2, _ := url.Parse("http://example.com/url2")

	ch1 := &Channel{track: &m3u8.Track{
		Name:  "CNN HD",
		URI:   uri1,
		Attrs: map[string]string{"tvg-id": "cnn-hd"},
		Tags:  map[string]string{"quality": "HD"},
	}}
	ch2 := &Channel{track: &m3u8.Track{
		Name:  "CNN 4K",
		URI:   uri2,
		Attrs: map[string]string{"tvg-id": "cnn-4k"},
		Tags:  map[string]string{"quality": "4K"},
	}}

	store.Add(ch1)
	store.Add(ch2)

	processor.Apply(store)

	channels := store.All()
	for _, ch := range channels {
		quality, exists := ch.Tags()["quality"]
		if !exists {
			t.Errorf("Expected quality tag to exist")
			continue
		}
		if quality != "CNN Multi" {
			t.Errorf("Expected quality tag 'CNN Multi', got '%s'", quality)
		}

		tvgID, exists := ch.GetAttr("tvg-id")
		if !exists {
			t.Errorf("Expected tvg-id attribute to exist")
			continue
		}
		if tvgID != "cnn-4k" {
			t.Errorf("Expected tvg-id 'cnn-4k', got '%s'", tvgID)
		}
	}
}

func mustTemplate(tmpl string) *common.Template {
	var t common.Template
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: tmpl,
	}
	if err := t.UnmarshalYAML(node); err != nil {
		panic(err)
	}
	return &t
}
