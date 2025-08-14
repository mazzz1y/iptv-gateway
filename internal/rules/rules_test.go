package rules_test

import (
	"regexp"
	"testing"
	"text/template"

	"iptv-gateway/internal/config"
	"iptv-gateway/internal/parser/m3u8"
	"iptv-gateway/internal/rules"

	"github.com/Masterminds/sprig/v3"
	"github.com/stretchr/testify/assert"
)

func tpl(name, s string) *template.Template {
	t, _ := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(s)
	return t
}

func TestRulesEngine_RemoveField(t *testing.T) {
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
								Value: config.RegexpArr{*regexp.MustCompile("^unwanted$")},
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
								Value: config.RegexpArr{*regexp.MustCompile("^test$")},
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
			rulesEngine := rules.NewRulesEngine(tt.rules)
			result := rulesEngine.Process(tt.track)

			assert.Equal(t, tt.shouldRemove, result)
			if !tt.shouldRemove && tt.expectedTrack != nil {
				assert.Equal(t, tt.expectedTrack, tt.track)
			}
		})
	}
}

func TestRulesEngine_SetField_MoveEquivalents(t *testing.T) {
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
								Value: config.RegexpArr{*regexp.MustCompile("^music$")},
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
								Value: config.RegexpArr{*regexp.MustCompile(".*")},
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
					"EXTGRP": "Entertainment",
				},
			},
			expectedTrack: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"group-name": "Entertainment",
				},
				Tags: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesEngine := rules.NewRulesEngine(tt.rules)
			result := rulesEngine.Process(tt.track)

			assert.False(t, result)
			assert.Equal(t, tt.expectedTrack, tt.track)
		})
	}
}

func TestRulesEngine_SetField_ReplaceEquivalents(t *testing.T) {
	tests := []struct {
		name          string
		rules         []config.RuleAction
		track         *m3u8.Track
		expectedTrack *m3u8.Track
	}{
		{
			name: "replace attr values",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Attr: &config.AttributeCondition{
								Name:  "tvg-group",
								Value: config.RegexpArr{*regexp.MustCompile("^music$")},
							},
						},
					},
					SetField: []config.SetFieldSpec{
						{
							Type:     "attr",
							Name:     "tvg-group",
							Template: tpl("attr:tvg-group", `{{ regexReplaceAll "^music$" (index .Channel.Attrs "tvg-group") "Entertainment" }}`),
						},
					},
				},
			},
			track: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-group": "music",
					"tvg-name":  "Music Channel",
				},
			},
			expectedTrack: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-group": "Entertainment",
					"tvg-name":  "Music Channel",
				},
			},
		},
		{
			name: "replace tag values",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Tag: &config.TagCondition{
								Name:  "EXTGRP",
								Value: config.RegexpArr{*regexp.MustCompile(".*")},
							},
						},
					},
					SetField: []config.SetFieldSpec{
						{
							Type:     "tag",
							Name:     "EXTGRP",
							Template: tpl("tag:EXTGRP", `{{ regexReplaceAll "^Group:" (index .Channel.Tags "EXTGRP") "" }}`),
						},
					},
				},
			},
			track: &m3u8.Track{
				Name: "Test Channel",
				Tags: map[string]string{
					"EXTGRP": "Group:Old Format",
				},
			},
			expectedTrack: &m3u8.Track{
				Name: "Test Channel",
				Tags: map[string]string{
					"EXTGRP": "Old Format",
				},
			},
		},
		{
			name: "replace without condition",
			rules: []config.RuleAction{
				{
					SetField: []config.SetFieldSpec{
						{
							Type:     "attr",
							Name:     "tvg-name",
							Template: tpl("attr:tvg-name", `{{ regexReplaceAll "\\s+" (index .Channel.Attrs "tvg-name") " " }}`),
						},
					},
				},
			},
			track: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-name": "  Music   Channel  ",
				},
			},
			expectedTrack: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-name": " Music Channel ",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesEngine := rules.NewRulesEngine(tt.rules)
			result := rulesEngine.Process(tt.track)

			assert.False(t, result)
			assert.Equal(t, tt.expectedTrack, tt.track)
		})
	}
}

func TestRulesEngine_SetField_CopyEquivalents(t *testing.T) {
	tests := []struct {
		name          string
		rules         []config.RuleAction
		track         *m3u8.Track
		expectedTrack *m3u8.Track
	}{
		{
			name: "copy attr to attr",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Attr: &config.AttributeCondition{
								Name:  "tvg-group",
								Value: config.RegexpArr{*regexp.MustCompile("^important$")},
							},
						},
					},
					SetField: []config.SetFieldSpec{
						{
							Type:     "attr",
							Name:     "backup-name",
							Template: tpl("attr:backup-name", `{{ index .Channel.Attrs "tvg-name" }}`),
						},
					},
				},
			},
			track: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-group": "important",
					"tvg-name":  "premium",
				},
			},
			expectedTrack: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-group":   "important",
					"tvg-name":    "premium",
					"backup-name": "premium",
				},
			},
		},
		{
			name: "copy name to attr",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Name: config.RegexpArr{*regexp.MustCompile(".*Channel.*")},
						},
					},
					SetField: []config.SetFieldSpec{
						{
							Type:     "attr",
							Name:     "channel-backup",
							Template: tpl("attr:channel-backup", `{{ .Channel.Name }}`),
						},
					},
				},
			},
			track: &m3u8.Track{
				Name: "Music Channel",
			},
			expectedTrack: &m3u8.Track{
				Name: "Music Channel",
				Attrs: map[string]string{
					"channel-backup": "Music Channel",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesEngine := rules.NewRulesEngine(tt.rules)
			result := rulesEngine.Process(tt.track)

			assert.False(t, result)
			assert.Equal(t, tt.expectedTrack, tt.track)
		})
	}
}

func TestRulesEngine_RemoveChannel(t *testing.T) {
	tests := []struct {
		name         string
		rules        []config.RuleAction
		track        *m3u8.Track
		shouldRemove bool
	}{
		{
			name: "remove channel by condition",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Attr: &config.AttributeCondition{
								Name:  "tvg-group",
								Value: config.RegexpArr{*regexp.MustCompile("^blocked$")},
							},
						},
					},
					RemoveChannel: &config.RemoveChannelRule{},
				},
			},
			track: &m3u8.Track{
				Name: "Test Channel",
				Attrs: map[string]string{
					"tvg-group": "blocked",
				},
			},
			shouldRemove: true,
		},
		{
			name: "keep channel by condition mismatch",
			rules: []config.RuleAction{
				{
					When: []config.Condition{
						{
							Name: config.RegexpArr{*regexp.MustCompile("^Other$")},
						},
					},
					RemoveChannel: &config.RemoveChannelRule{},
				},
			},
			track: &m3u8.Track{
				Name: "Test Channel",
			},
			shouldRemove: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesEngine := rules.NewRulesEngine(tt.rules)
			result := rulesEngine.Process(tt.track)
			assert.Equal(t, tt.shouldRemove, result)
		})
	}
}
