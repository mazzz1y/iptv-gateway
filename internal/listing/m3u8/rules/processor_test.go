package rules

import (
	"net/url"
	"regexp"
	"testing"

	"iptv-gateway/internal/config/common"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/urlgen"
)

func mustCompile(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

type mockPlaylist struct {
	name string
}

func (m mockPlaylist) Name() string                    { return m.name }
func (m mockPlaylist) Playlists() []string             { return nil }
func (m mockPlaylist) URLGenerator() *urlgen.Generator { return nil }
func (m mockPlaylist) Rules() []*rules.ChannelRule     { return nil }
func (m mockPlaylist) IsProxied() bool                 { return false }

func TestConditionLogic(t *testing.T) {
	playlist := mockPlaylist{name: "pl1"}
	uri, _ := url.Parse("http://example.com/stream")
	track := &m3u8.Track{
		Name: "Channel A",
		URI:  uri,
		Tags: map[string]string{"cat": "restricted"},
	}
	channel := NewChannel(track, playlist)
	processor := NewProcessor("client1", nil, nil)

	tests := []struct {
		condition   common.Condition
		expectMatch bool
	}{
		{
			condition: common.Condition{
				Clients:  common.StringOrArr{"client1", "client2"},
				Selector: &common.Selector{Type: common.SelectorTag, Value: "cat"},
				Patterns: common.RegexpArr{mustCompile("restricted")},
			},
			expectMatch: true,
		},
		{
			condition: common.Condition{
				Clients:  common.StringOrArr{"client1", "client2"},
				Selector: &common.Selector{Type: common.SelectorTag, Value: "cat"},
				Patterns: common.RegexpArr{mustCompile("safe")},
			},
			expectMatch: false,
		},
		{
			condition: common.Condition{
				Clients:  common.StringOrArr{"client3"},
				Selector: &common.Selector{Type: common.SelectorTag, Value: "cat"},
				Patterns: common.RegexpArr{mustCompile("restricted")},
			},
			expectMatch: false,
		},
		{
			condition: common.Condition{
				Clients: common.StringOrArr{"client1", "client2"},
			},
			expectMatch: true,
		},
		{
			condition: common.Condition{
				Selector: &common.Selector{Type: common.SelectorTag, Value: "cat"},
				Patterns: common.RegexpArr{mustCompile("restricted")},
			},
			expectMatch: true,
		},
	}

	for _, tt := range tests {
		result := processor.matchesCondition(channel, tt.condition)
		if result != tt.expectMatch {
			t.Errorf("matchesCondition() = %v, want %v", result, tt.expectMatch)
		}

		rule := &rules.RemoveChannelRule{Condition: &tt.condition}
		shouldRemove := processor.processRemoveChannel(channel, rule)
		if shouldRemove != tt.expectMatch {
			t.Errorf("processRemoveChannel() = %v, want %v", shouldRemove, tt.expectMatch)
		}
		channel.removed = false
	}
}

func TestPlaylistCondition(t *testing.T) {
	playlist := mockPlaylist{name: "pl2"}
	uri, _ := url.Parse("http://example.com/stream")
	track := &m3u8.Track{Name: "Channel B", URI: uri}
	channel := NewChannel(track, playlist)
	processor := NewProcessor("client1", nil, nil)

	condition := common.Condition{
		Clients:   common.StringOrArr{"client1", "client2"},
		Playlists: common.StringOrArr{"pl2"},
	}

	result := processor.matchesCondition(channel, condition)
	if !result {
		t.Error("Expected match when both client and playlist match")
	}

	condition.Playlists = common.StringOrArr{"pl3"}
	result = processor.matchesCondition(channel, condition)
	if result {
		t.Error("Expected no match when playlist doesn't match")
	}
}

func TestAdultChannelFilteringWithClientAndOrConditions(t *testing.T) {
	playlist := mockPlaylist{name: "test-playlist"}
	uri, _ := url.Parse("http://example.com/stream")

	track := &m3u8.Track{
		Name: "NSFW Adult Channel",
		Tags: map[string]string{
			"EXTGRP": "adult",
		},
		URI: uri,
	}
	channel := NewChannel(track, playlist)

	restrictedProcessor := NewProcessor("tv-bedroom", nil, nil)
	condition := common.Condition{
		Clients: common.StringOrArr{"tv-bedroom", "tv2-bedroom"},
		Or: []common.Condition{
			{
				Selector: &common.Selector{Type: common.SelectorName},
				Patterns: common.RegexpArr{mustCompile(".*NSFW.*")},
			},
			{
				Selector: &common.Selector{Type: common.SelectorTag, Value: "EXTGRP"},
				Patterns: common.RegexpArr{mustCompile("(?i)adult")},
			},
		},
	}

	result := restrictedProcessor.matchesCondition(channel, condition)
	if !result {
		t.Error("Expected match for restricted client with adult content")
	}

	allowedProcessor := NewProcessor("living-room", nil, nil)
	result = allowedProcessor.matchesCondition(channel, condition)
	if result {
		t.Error("Expected no match for non-restricted client - adult content should be available")
	}

	restrictedProcessor2 := NewProcessor("tv2-bedroom", nil, nil)
	result = restrictedProcessor2.matchesCondition(channel, condition)
	if !result {
		t.Error("Expected match for tv2-bedroom restricted client with adult content")
	}
}

func TestEvaluateFieldConditionEdgeCases(t *testing.T) {
	playlist := mockPlaylist{name: "test-playlist"}
	uri, _ := url.Parse("http://example.com/stream")
	track := &m3u8.Track{Name: "Test Channel", URI: uri}
	channel := NewChannel(track, playlist)
	processor := NewProcessor("client1", nil, nil)

	emptyCondition := common.Condition{}
	result := processor.evaluateConditionFieldCondition(channel, emptyCondition)
	if !result {
		t.Error("Expected true for empty field condition")
	}

	conditionWithOnlyOr := common.Condition{
		Or: []common.Condition{
			{
				Selector: &common.Selector{Type: common.SelectorName},
				Patterns: common.RegexpArr{mustCompile("Test.*")},
			},
		},
	}
	result = processor.evaluateConditionFieldCondition(channel, conditionWithOnlyOr)
	if !result {
		t.Error("Expected true when no field conditions are specified")
	}

	conditionMissingAttr := common.Condition{
		Selector: &common.Selector{Type: common.SelectorAttr, Value: "non-existent-attr"},
		Patterns: common.RegexpArr{mustCompile(".*")},
	}
	result = processor.evaluateConditionFieldCondition(channel, conditionMissingAttr)
	if result {
		t.Error("Expected false for non-existent attribute")
	}

	conditionMissingTag := common.Condition{
		Selector: &common.Selector{Type: common.SelectorTag, Value: "non-existent-tag"},
		Patterns: common.RegexpArr{mustCompile(".*")},
	}
	result = processor.evaluateConditionFieldCondition(channel, conditionMissingTag)
	if result {
		t.Error("Expected false for non-existent tag")
	}
}
