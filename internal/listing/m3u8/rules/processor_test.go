package rules

import (
	"net/url"
	"regexp"
	"testing"

	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
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
		condition   types.Condition
		expectMatch bool
	}{
		{
			condition: types.Condition{
				Clients: types.StringOrArr{"client1", "client2"},
				Tag: &types.NamePatterns{
					Name:     "cat",
					Patterns: types.RegexpArr{mustCompile("restricted")},
				},
			},
			expectMatch: true,
		},
		{
			condition: types.Condition{
				Clients: types.StringOrArr{"client1", "client2"},
				Tag: &types.NamePatterns{
					Name:     "cat",
					Patterns: types.RegexpArr{mustCompile("safe")},
				},
			},
			expectMatch: false,
		},
		{
			condition: types.Condition{
				Clients: types.StringOrArr{"client3"},
				Tag: &types.NamePatterns{
					Name:     "cat",
					Patterns: types.RegexpArr{mustCompile("restricted")},
				},
			},
			expectMatch: false,
		},
		{
			condition: types.Condition{
				Clients: types.StringOrArr{"client1", "client2"},
			},
			expectMatch: true,
		},
		{
			condition: types.Condition{
				Tag: &types.NamePatterns{
					Name:     "cat",
					Patterns: types.RegexpArr{mustCompile("restricted")},
				},
			},
			expectMatch: true,
		},
	}

	for _, tt := range tests {
		result := processor.matchesCondition(channel, tt.condition)
		if result != tt.expectMatch {
			t.Errorf("matchesCondition() = %v, want %v", result, tt.expectMatch)
		}

		rule := &rules.RemoveChannelRule{When: &tt.condition}
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

	condition := types.Condition{
		Clients:   types.StringOrArr{"client1", "client2"},
		Playlists: types.StringOrArr{"pl2"},
	}

	result := processor.matchesCondition(channel, condition)
	if !result {
		t.Error("Expected match when both client and playlist match")
	}

	condition.Playlists = types.StringOrArr{"pl3"}
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
	condition := types.Condition{
		Clients: types.StringOrArr{"tv-bedroom", "tv2-bedroom"},
		Or: []types.Condition{
			{
				NamePatterns: types.RegexpArr{mustCompile(".*NSFW.*")},
			},
			{
				Tag: &types.NamePatterns{
					Name:     "EXTGRP",
					Patterns: types.RegexpArr{mustCompile("(?i)adult")},
				},
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

	emptyCondition := types.Condition{}
	result := processor.evaluateFieldCondition(channel, emptyCondition)
	if !result {
		t.Error("Expected true for empty field condition")
	}

	conditionWithOnlyOr := types.Condition{
		Or: []types.Condition{
			{NamePatterns: types.RegexpArr{mustCompile("Test.*")}},
		},
	}
	result = processor.evaluateFieldCondition(channel, conditionWithOnlyOr)
	if !result {
		t.Error("Expected true when no field conditions are specified")
	}

	conditionMissingAttr := types.Condition{
		Attr: &types.NamePatterns{
			Name:     "non-existent-attr",
			Patterns: types.RegexpArr{mustCompile(".*")},
		},
	}
	result = processor.evaluateFieldCondition(channel, conditionMissingAttr)
	if result {
		t.Error("Expected false for non-existent attribute")
	}

	conditionMissingTag := types.Condition{
		Tag: &types.NamePatterns{
			Name:     "non-existent-tag",
			Patterns: types.RegexpArr{mustCompile(".*")},
		},
	}
	result = processor.evaluateFieldCondition(channel, conditionMissingTag)
	if result {
		t.Error("Expected false for non-existent tag")
	}
}
