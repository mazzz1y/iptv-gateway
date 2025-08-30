package rules

import (
	"bytes"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/parser/m3u8"
)

type Processor struct {
	subscriptionRulesMap map[Subscription][]rules.RuleAction
}

func NewProcessor() *Processor {
	return &Processor{
		subscriptionRulesMap: make(map[Subscription][]rules.RuleAction),
	}
}

func (p *Processor) AddSubscriptionRules(sub Subscription, rules []rules.RuleAction) {
	p.subscriptionRulesMap[sub] = append(p.subscriptionRulesMap[sub], rules...)
}

func (p *Processor) Process(store *Store) {
	p.processStoreRules(store)
	p.processTrackRules(store)
}

func (p *Processor) processStoreRules(store *Store) {
	p.processRemoveDuplicatesRules(store)
}

func (p *Processor) processTrackRules(store *Store) {
	for _, ch := range store.All() {
		track := ch.Track()

		if subRules, exists := p.subscriptionRulesMap[ch.Subscription()]; exists {
			p.processTrackWithRules(track, subRules)
		}
	}
}

func (p *Processor) processTrackWithRules(track *m3u8.Track, rules []rules.RuleAction) {
	for _, action := range rules {
		if !p.matchesConditions(track, action.When) {
			continue
		}

		if action.RemoveChannel != nil {
			track.IsRemoved = true
			return
		}

		if action.SetField != nil {
			p.setFields(track, action.SetField)
		}

		if action.RemoveField != nil {
			p.removeFields(track, action.RemoveField)
		}
	}
}

func (p *Processor) processRemoveDuplicatesRules(global *Store) {
	for sub, rules := range p.subscriptionRulesMap {
		subStore := NewStore()
		for _, ch := range global.All() {
			if ch.Subscription() == sub {
				subStore.Add(ch)
			}
		}

		for _, action := range rules {
			if action.RemoveChannelDups != nil {
				for _, dupRule := range *action.RemoveChannelDups {
					rule := NewRemoveDuplicatesRule(dupRule.Patterns, dupRule.TrimPattern)
					rule.Apply(global, subStore)
				}
			}
		}
	}
}

func (p *Processor) matchesConditions(track *m3u8.Track, conditions []rules.Condition) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, condition := range conditions {
		if !p.matchesCondition(track, condition) {
			return false
		}
	}
	return true
}

func (p *Processor) matchesCondition(track *m3u8.Track, condition rules.Condition) bool {
	if condition.IsEmpty() {
		return true
	}

	if len(condition.Name) > 0 {
		if !p.matchesRegexps(track.Name, condition.Name) {
			return false
		}
	}

	if condition.Attr != nil {
		var actual string
		var exists bool
		if track.Attrs != nil {
			actual, exists = track.Attrs[condition.Attr.Name]
		}
		if !exists || !p.matchesRegexps(actual, condition.Attr.Value) {
			return false
		}
	}

	if condition.Tag != nil {
		var actual string
		var exists bool
		if track.Tags != nil {
			actual, exists = track.Tags[condition.Tag.Name]
		}
		if !exists || !p.matchesRegexps(actual, condition.Tag.Value) {
			return false
		}
	}

	if len(condition.And) > 0 {
		for _, subCondition := range condition.And {
			if !p.matchesCondition(track, subCondition) {
				return false
			}
		}
	}

	if len(condition.Or) > 0 {
		for _, subCondition := range condition.Or {
			if p.matchesCondition(track, subCondition) {
				return true
			}
		}
		return false
	}

	if len(condition.Not) > 0 {
		for _, subCondition := range condition.Not {
			if p.matchesCondition(track, subCondition) {
				return false
			}
		}
	}

	return true
}

func (p *Processor) matchesRegexps(value string, regexps types.RegexpArr) bool {
	for _, re := range regexps {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func (p *Processor) removeFields(track *m3u8.Track, fields []map[string]types.RegexpArr) {
	for _, field := range fields {
		for fieldType, regexps := range field {
			switch fieldType {
			case "attr":
				if track.Attrs != nil {
					for attrName := range track.Attrs {
						if p.matchesRegexps(attrName, regexps) {
							delete(track.Attrs, attrName)
						}
					}
				}
			case "tag":
				if track.Tags != nil {
					for tagName := range track.Tags {
						if p.matchesRegexps(tagName, regexps) {
							delete(track.Tags, tagName)
						}
					}
				}
			case "name":
				if p.matchesRegexps("name", regexps) {
					track.Name = ""
				}
			}
		}
	}
}

func (p *Processor) setFields(track *m3u8.Track, fields []rules.SetFieldSpec) {
	for _, field := range fields {
		var buf bytes.Buffer
		_ = field.Template.Execute(&buf, map[string]any{
			"Channel": map[string]any{
				"Name":  track.Name,
				"Attrs": track.Attrs,
				"Tags":  track.Tags,
			},
		})
		p.setValue(track, field.Type, field.Name, buf.String())
	}
}

func (p *Processor) setValue(track *m3u8.Track, fieldType, fieldName, value string) {
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
