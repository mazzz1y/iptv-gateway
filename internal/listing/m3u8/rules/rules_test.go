package rules_test

import (
	"iptv-gateway/internal/listing/m3u8/rules"
	"regexp"
	"testing"
	"text/template"

	"iptv-gateway/internal/config"
	"iptv-gateway/internal/parser/m3u8"

	"github.com/Masterminds/sprig/v3"
	"github.com/stretchr/testify/assert"
)

func tpl(name, s string) *template.Template {
	t, _ := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(s)
	return t
}

type mockSubscription struct {
	name string
}

func (m mockSubscription) IsProxied() bool {
	return false
}

func TestRulesProcessor_RemoveField(t *testing.T) {
	tests := []struct {
		name          string
		rules         []config.RuleAction
		track         *m3u8.Track
		shouldRemove  bool
		expectedTrack *m3u8.Track
	}{
		{
			name: "remove channel by attr",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Attr: &config.AttributeCondition{
								Name:  "tvg-group",
								Value: config.RegexpArr{regexp.MustCompile("^unwanted$")},
							},
						},
					},
					RemoveChannel: &config.RemoveChannelRule{},
				},
			},
			track: &m3u8.Track{
				Name:  "Test Channel",
				Attrs: map[string]string{"tvg-group": "unwanted"},
			},
			shouldRemove: true,
		},
		{
			name: "remove fields",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Attr: &config.AttributeCondition{
								Name:  "tvg-group",
								Value: config.RegexpArr{regexp.MustCompile("^test$")},
							},
						},
					},
					RemoveField: []config.FieldSpec{
						{Type: "attr", Name: "tvg-id"},
						{Type: "tag", Name: "EXTBYT"},
					},
				},
			},
			track: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-group": "test",
					"tvg-id":    "123",
					"tvg-name":  "Channel",
				},
				Tags: map[string]string{
					"EXTBYT": "data",
					"EXTGRP": "group",
				},
			},
			shouldRemove: false,
			expectedTrack: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-group": "test",
					"tvg-name":  "Channel",
				},
				Tags: map[string]string{
					"EXTGRP": "group",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := rules.NewProcessor()
			sub := mockSubscription{name: "test"}
			processor.AddSubscriptionRules(sub, tt.rules)

			store := rules.NewStore()
			channel := rules.NewChannel(tt.track, sub)
			store.Add(channel)

			processor.Process(store)

			assert.Equal(t, tt.shouldRemove, tt.track.IsRemoved)
			if !tt.shouldRemove && tt.expectedTrack != nil {
				assert.Equal(t, tt.expectedTrack.Name, tt.track.Name)
				assert.Equal(t, tt.expectedTrack.Attrs, tt.track.Attrs)
				assert.Equal(t, tt.expectedTrack.Tags, tt.track.Tags)
			}
		})
	}
}

func TestRulesProcessor_SetField_MoveEquivalents(t *testing.T) {
	tests := []struct {
		name          string
		rules         []config.RuleAction
		track         *m3u8.Track
		expectedTrack *m3u8.Track
	}{
		{
			name: "move attr to tag",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Attr: &config.AttributeCondition{
								Name:  "tvg-group",
								Value: config.RegexpArr{regexp.MustCompile("^music$")},
							},
						},
					},
					SetField: []config.SetFieldSpec{
						{
							Type:     "tag",
							Name:     "EXTGRP",
							Template: tpl("tag:EXTGRP", `{{ index .Channel.Attrs "tvg-group" }}`),
						},
						{
							Type:     "tag",
							Name:     "EXT-X-LOGO",
							Template: tpl("tag:EXT-X-LOGO", `{{ index .Channel.Attrs "tvg-logo" }}`),
						},
					},
					RemoveField: []config.FieldSpec{
						{Type: "attr", Name: "tvg-group"},
						{Type: "attr", Name: "tvg-logo"},
					},
				},
			},
			track: &m3u8.Track{
				Name: "Music Channel",
				Attrs: map[string]string{
					"tvg-group": "music",
					"tvg-logo":  "http://example.com/logo.png",
					"tvg-name":  "Music Channel",
				},
			},
			expectedTrack: &m3u8.Track{
				Name: "Music Channel",
				Attrs: map[string]string{
					"tvg-name": "Music Channel",
				},
				Tags: map[string]string{
					"EXTGRP":     "music",
					"EXT-X-LOGO": "http://example.com/logo.png",
				},
			},
		},
		{
			name: "move tag to attr",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Tag: &config.TagCondition{
								Name:  "EXTGRP",
								Value: config.RegexpArr{regexp.MustCompile(".*")},
							},
						},
					},
					SetField: []config.SetFieldSpec{
						{
							Type:     "attr",
							Name:     "group-name",
							Template: tpl("attr:group-name", `{{ index .Channel.Tags "EXTGRP" }}`),
						},
					},
					RemoveField: []config.FieldSpec{
						{Type: "tag", Name: "EXTGRP"},
					},
				},
			},
			track: &m3u8.Track{
				Name: "Test Channel",
				Tags: map[string]string{
					"EXTGRP": "entertainment",
				},
			},
			expectedTrack: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"group-name": "entertainment",
				},
				Tags: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := rules.NewProcessor()
			sub := mockSubscription{name: "test"}
			processor.AddSubscriptionRules(sub, tt.rules)

			store := rules.NewStore()
			channel := rules.NewChannel(tt.track, sub)
			store.Add(channel)

			processor.Process(store)

			assert.Equal(t, tt.expectedTrack.Name, tt.track.Name)
			assert.Equal(t, tt.expectedTrack.Attrs, tt.track.Attrs)
			assert.Equal(t, tt.expectedTrack.Tags, tt.track.Tags)
		})
	}
}
