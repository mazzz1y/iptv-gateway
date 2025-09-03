package rules_test

import (
	"regexp"
	"testing"
	"text/template"

	configrules "iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/shell"
	"iptv-gateway/internal/urlgen"

	"github.com/Masterminds/sprig/v3"
	"github.com/stretchr/testify/assert"
)

func tpl(name, s string) *template.Template {
	t, _ := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(s)
	return t
}

type mockSubscription struct {
	name         string
	channelRules []configrules.ChannelRule
}

func (m mockSubscription) IsProxied() bool {
	return false
}

func (m mockSubscription) Playlists() []string {
	return nil
}

func (m mockSubscription) URLGenerator() *urlgen.Generator {
	return nil
}

func (m mockSubscription) ChannelRules() []configrules.ChannelRule {
	return m.channelRules
}

func (m mockSubscription) PlaylistRules() []configrules.PlaylistRule {
	return nil
}

func (m mockSubscription) Name() string {
	return m.name
}

func (m mockSubscription) ExpiredCommandStreamer() *shell.Streamer {
	return nil
}

func TestRulesProcessor_RemoveField(t *testing.T) {
	tests := []struct {
		name          string
		rules         []configrules.ChannelRule
		track         *m3u8.Track
		shouldRemove  bool
		expectedTrack *m3u8.Track
	}{
		{
			name: "remove channel by attr",
			rules: []configrules.ChannelRule{
				{
					When: []configrules.Condition{
						{
							Attr: &configrules.AttributeCondition{
								Name:  "tvg-group",
								Value: types.RegexpArr{regexp.MustCompile("^unwanted$")},
							},
						},
					},
					RemoveChannel: func() *any { v := any(struct{}{}); return &v }(),
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
			rules: []configrules.ChannelRule{
				{
					When: []configrules.Condition{
						{
							Attr: &configrules.AttributeCondition{
								Name:  "tvg-group",
								Value: types.RegexpArr{regexp.MustCompile("^test$")},
							},
						},
					},
					RemoveField: []map[string]types.RegexpArr{
						{"attr": types.RegexpArr{regexp.MustCompile("tvg-id")}},
						{"tag": types.RegexpArr{regexp.MustCompile("EXTBYT")}},
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
			sub := &mockSubscription{name: "test", channelRules: tt.rules}
			processor.AddSubscription(sub)

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
		rules         []configrules.ChannelRule
		track         *m3u8.Track
		expectedTrack *m3u8.Track
	}{
		{
			name: "move attr to tag",
			rules: []configrules.ChannelRule{
				{
					When: []configrules.Condition{
						{
							Attr: &configrules.AttributeCondition{
								Name:  "tvg-group",
								Value: types.RegexpArr{regexp.MustCompile("^music$")},
							},
						},
					},
					SetField: []configrules.SetFieldSpec{
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
					RemoveField: []map[string]types.RegexpArr{
						{"attr": types.RegexpArr{regexp.MustCompile("tvg-group")}},
						{"attr": types.RegexpArr{regexp.MustCompile("tvg-logo")}},
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
			rules: []configrules.ChannelRule{
				{
					When: []configrules.Condition{
						{
							Tag: &configrules.TagCondition{
								Name:  "EXTGRP",
								Value: types.RegexpArr{regexp.MustCompile(".*")},
							},
						},
					},
					SetField: []configrules.SetFieldSpec{
						{
							Type:     "attr",
							Name:     "group-name",
							Template: tpl("attr:group-name", `{{ index .Channel.Tags "EXTGRP" }}`),
						},
					},
					RemoveField: []map[string]types.RegexpArr{
						{"tag": types.RegexpArr{regexp.MustCompile("EXTGRP")}},
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
			sub := &mockSubscription{name: "test", channelRules: tt.rules}
			processor.AddSubscription(sub)

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
