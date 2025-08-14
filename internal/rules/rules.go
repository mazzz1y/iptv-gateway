package rules

import (
	"bytes"
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/parser/m3u8"
	"regexp"
)

type Engine struct {
	rules []config.RuleAction
}

func NewRulesEngine(rules []config.RuleAction) *Engine {
	return &Engine{
		rules: rules,
	}
}

func (e *Engine) Process(track *m3u8.Track) bool {
	for _, action := range e.rules {
		if shouldSkip := e.applyAction(track, action); shouldSkip {
			return true
		}
	}
	return false
}

func (e *Engine) applyAction(track *m3u8.Track, action config.RuleAction) bool {
	if !e.matchesConditions(track, action.When) {
		return false
	}

	if action.RemoveChannel != nil {
		return true
	}

	if action.SetField != nil {
		e.setFields(track, action.SetField)
	}

	if action.RemoveField != nil {
		e.removeFields(track, action.RemoveField)
	}

	return false
}

func (e *Engine) matchesConditions(track *m3u8.Track, conditions []config.Condition) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, condition := range conditions {
		if !e.matchesCondition(track, condition) {
			return false
		}
	}
	return true
}

func (e *Engine) matchesCondition(track *m3u8.Track, condition config.Condition) bool {
	if condition.IsEmpty() {
		return true
	}

	if len(condition.Name) > 0 {
		if !e.matchesRegexps(track.Name, condition.Name) {
			return false
		}
	}

	if condition.Attr != nil {
		var actual string
		var exists bool
		if track.Attrs != nil {
			actual, exists = track.Attrs[condition.Attr.Name]
		}
		if !exists || !e.matchesRegexps(actual, condition.Attr.Value) {
			return false
		}
	}

	if condition.Tag != nil {
		var actual string
		var exists bool
		if track.Tags != nil {
			actual, exists = track.Tags[condition.Tag.Name]
		}
		if !exists || !e.matchesRegexps(actual, condition.Tag.Value) {
			return false
		}
	}

	if len(condition.And) > 0 {
		for _, subCondition := range condition.And {
			if !e.matchesCondition(track, subCondition) {
				return false
			}
		}
	}

	if len(condition.Or) > 0 {
		for _, subCondition := range condition.Or {
			if e.matchesCondition(track, subCondition) {
				return true
			}
		}
		return false
	}

	return true
}

func (e *Engine) matchesRegexps(value string, regexps []regexp.Regexp) bool {
	for _, re := range regexps {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func (e *Engine) removeFields(track *m3u8.Track, fields []config.FieldSpec) {
	for _, field := range fields {
		switch field.Type {
		case "attr":
			if track.Attrs != nil {
				delete(track.Attrs, field.Name)
			}
		case "tag":
			if track.Tags != nil {
				delete(track.Tags, field.Name)
			}
		case "name":
			track.Name = ""
		}
	}
}

func (e *Engine) setFields(track *m3u8.Track, fields []config.SetFieldSpec) {
	for _, field := range fields {
		var buf bytes.Buffer
		_ = field.Template.Execute(&buf, map[string]any{
			"Channel": map[string]any{
				"Name":  track.Name,
				"Attrs": track.Attrs,
				"Tags":  track.Tags,
			},
		})
		e.setValue(track, field.Type, field.Name, buf.String())
	}
}

func (e *Engine) setValue(track *m3u8.Track, fieldType, fieldName, value string) {
	switch fieldType {
	case "attr":
		if track.Attrs == nil {
			track.Attrs = make(map[string]string)
		}
		track.Attrs[fieldName] = value
	case "tag":
		if track.Tags == nil {
			track.Tags = make(map[string]string)
		}
		track.Tags[fieldName] = value
	case "name":
		track.Name = value
	}
}
