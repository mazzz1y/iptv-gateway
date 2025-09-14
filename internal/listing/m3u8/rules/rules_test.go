package rules_test

import (
	"bytes"
	configrules "iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/rules/channel"
	"iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/listing/m3u8/rules"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/shell"
	"iptv-gateway/internal/urlgen"
	"regexp"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type mockSubscription struct {
	name         string
	channelRules []channel.Rule
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

func (m mockSubscription) ChannelRules() []channel.Rule {
	return m.channelRules
}

func (m mockSubscription) PlaylistRules() []playlist.Rule {
	return nil
}

func (m mockSubscription) NamedConditions() []configrules.Condition {
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
		rules         []channel.Rule
		track         *m3u8.Track
		shouldRemove  bool
		expectedTrack *m3u8.Track
	}{
		{
			name: "remove channel by attr",
			rules: []channel.Rule{
				{
					RemoveChannel: &channel.RemoveChannelRule{
						When: &configrules.Condition{
							Attr: &types.NamePatterns{
								Name:     "tvg-group",
								Patterns: types.RegexpArr{mustCompileRegexp("unwanted")},
							},
						},
					},
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
			rules: []channel.Rule{
				{
					SetField: &channel.SetFieldRule{
						AttrTemplate: &types.NameTemplate{
							Name:     "tvg-id",
							Template: mustCreateTemplate(""),
						},
						When: &configrules.Condition{
							Attr: &types.NamePatterns{
								Name:     "tvg-group",
								Patterns: types.RegexpArr{mustCompileRegexp("test")},
							},
						},
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
					"tvg-id":    "",
					"tvg-name":  "Channel",
				},
				Tags: map[string]string{
					"EXTBYT": "data",
					"EXTGRP": "group",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := rules.NewProcessor(nil)
			sub := &mockSubscription{name: "test", channelRules: tt.rules}
			processor.AddSubscription(sub)

			store := rules.NewStore()
			ch := rules.NewChannel(tt.track, sub)
			store.Add(ch)

			processor.Process(store)

			assert.Equal(t, tt.shouldRemove, ch.IsRemoved())
			if !tt.shouldRemove && tt.expectedTrack != nil {
				assert.Equal(t, tt.expectedTrack.Name, tt.track.Name)
				assert.Equal(t, tt.expectedTrack.Attrs, tt.track.Attrs)
				assert.Equal(t, tt.expectedTrack.Tags, tt.track.Tags)
			}
		})
	}
}

func TestRulesProcessor_SetField(t *testing.T) {
	channelRules := []channel.Rule{
		{
			SetField: &channel.SetFieldRule{
				AttrTemplate: &types.NameTemplate{
					Name:     "tvg-group",
					Template: mustCreateTemplate("music"),
				},
			},
		},
	}

	processor := rules.NewProcessor(nil)
	sub := &mockSubscription{name: "test", channelRules: channelRules}
	processor.AddSubscription(sub)

	track := &m3u8.Track{
		Name: "Test Channel",
		Attrs: map[string]string{
			"tvg-name": "Test Channel",
		},
	}

	store := rules.NewStore()
	ch := rules.NewChannel(track, sub)
	store.Add(ch)

	processor.Process(store)

	assert.Equal(t, "music", track.Attrs["tvg-group"])
}

func TestTemplate(t *testing.T) {
	tmplMap := map[string]any{
		"Channel": map[string]any{
			"NamePatterns": "Test Channel",
			"Attrs":        map[string]string{"tvg-group": "movies", "tvg-id": "123"},
			"Tags":         map[string]string{"EXTBYT": "data"},
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "simple text",
			template: "music",
			expected: "music",
		},
		{
			name:     "channel name",
			template: "{{ .Channel.NamePatterns }}",
			expected: "Test Channel",
		},
		{
			name:     "channel attr",
			template: `{{ index .Channel.Attrs "tvg-group" }}`,
			expected: "movies",
		},
		{
			name:     "channel tag",
			template: "{{ .Channel.Tags.EXTBYT }}",
			expected: "data",
		},
		{
			name:     "combined template",
			template: `{{ .Channel.NamePatterns }}-{{ index .Channel.Attrs "tvg-group" }}`,
			expected: "Test Channel-movies",
		},
		{
			name:     "with sprig functions",
			template: "{{ .Channel.NamePatterns | upper }}",
			expected: "TEST CHANNEL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := mustCreateTemplate(tt.template)
			goTmpl := tmpl.ToTemplate()

			var buf bytes.Buffer
			err := goTmpl.Execute(&buf, tmplMap)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func mustCreateTemplate(text string) *types.Template {
	tmpl, err := template.New("test").Funcs(sprig.FuncMap()).Parse(text)
	if err != nil {
		panic(err)
	}
	result := types.Template(*tmpl)
	return &result
}

func mustCompileRegexp(pattern string) *regexp.Regexp {
	result := &types.RegexpArr{}
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: pattern,
	}
	if err := result.UnmarshalYAML(node); err != nil {
		panic(err)
	}
	if len(*result) == 0 {
		panic("no regexp compiled")
	}
	return (*result)[0]
}

type yamlNode struct {
	Value string
}

func (n *yamlNode) Decode(v interface{}) error {
	if s, ok := v.(*string); ok {
		*s = n.Value
		return nil
	}
	return nil
}
