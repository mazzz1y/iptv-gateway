package rules

import (
	"bytes"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/listing"
)

type Processor struct {
	subscriptionChannelRulesMap  map[listing.Playlist][]rules.ChannelRule
	subscriptionPlaylistRulesMap map[listing.Playlist][]rules.PlaylistRule
}

func NewProcessor() *Processor {
	return &Processor{
		subscriptionChannelRulesMap:  make(map[listing.Playlist][]rules.ChannelRule),
		subscriptionPlaylistRulesMap: make(map[listing.Playlist][]rules.PlaylistRule),
	}
}

func (p *Processor) AddSubscription(sub listing.Playlist) {
	expandedChannelRules := p.expandNamedConditions(sub.ChannelRules(), sub.NamedConditions())
	p.subscriptionChannelRulesMap[sub] = append(p.subscriptionChannelRulesMap[sub], expandedChannelRules...)
	p.subscriptionPlaylistRulesMap[sub] = append(p.subscriptionPlaylistRulesMap[sub], sub.PlaylistRules()...)
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
		if subRules, exists := p.subscriptionChannelRulesMap[ch.Subscription()]; exists {
			p.processChannelWithRules(ch, subRules)
		}
	}
}

func (p *Processor) processChannelWithRules(ch *Channel, rules []rules.ChannelRule) {
	for _, action := range rules {
		if !p.matchesConditions(ch, action.When) {
			continue
		}

		if action.RemoveChannel != nil {
			ch.MarkRemoved()
			return
		}

		if action.MarkHidden != nil {
			ch.MarkHidden()
		}

		if action.SetField != nil {
			p.setFields(ch, action.SetField)
		}

		if action.RemoveField != nil {
			p.removeFields(ch, action.RemoveField)
		}
	}
}

func (p *Processor) processRemoveDuplicatesRules(global *Store) {
	for sub, rul := range p.subscriptionPlaylistRulesMap {
		subStore := NewStore()
		for _, ch := range global.All() {
			if ch.Subscription() == sub {
				subStore.Add(ch)
			}
		}

		for _, action := range rul {
			if action.RemoveDuplicates != nil {
				for _, dupRule := range *action.RemoveDuplicates {
					rule := NewRemoveDuplicatesRule(dupRule.Patterns, dupRule.TrimPattern)
					rule.Apply(global, subStore)
				}
			}
		}
	}
}

func (p *Processor) matchesConditions(ch *Channel, conditions []rules.Condition) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, condition := range conditions {
		if p.matchesCondition(ch, condition) {
			return true
		}
	}
	return false
}

func (p *Processor) matchesCondition(ch *Channel, condition rules.Condition) bool {
	if condition.IsEmpty() {
		return true
	}

	if len(condition.Name) > 0 {
		if !p.matchesRegexps(ch.Name(), condition.Name) {
			return false
		}
	}

	if condition.Attr != nil {
		actual, exists := ch.GetAttr(condition.Attr.Name)
		if !exists || !p.matchesRegexps(actual, condition.Attr.Value) {
			return false
		}
	}

	if condition.Tag != nil {
		actual, exists := ch.GetTag(condition.Tag.Name)
		if !exists || !p.matchesRegexps(actual, condition.Tag.Value) {
			return false
		}
	}

	if len(condition.And) > 0 {
		for _, subCondition := range condition.And {
			if !p.matchesCondition(ch, subCondition) {
				return false
			}
		}
	}

	if len(condition.Or) > 0 {
		for _, subCondition := range condition.Or {
			if p.matchesCondition(ch, subCondition) {
				return true
			}
		}
		return false
	}

	if len(condition.Not) > 0 {
		anyMatched := false
		for _, subCondition := range condition.Not {
			if p.matchesCondition(ch, subCondition) {
				anyMatched = true
				break
			}
		}
		return !anyMatched
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

func (p *Processor) removeFields(ch *Channel, fields []map[string]types.RegexpArr) {
	for _, field := range fields {
		for fieldType, regexps := range field {
			switch fieldType {
			case "attr":
				if ch.Attrs() != nil {
					for attrName := range ch.Attrs() {
						if p.matchesRegexps(attrName, regexps) {
							ch.DeleteAttr(attrName)
						}
					}
				}
			case "tag":
				if ch.Tags() != nil {
					for tagName := range ch.Tags() {
						if p.matchesRegexps(tagName, regexps) {
							ch.DeleteTag(tagName)
						}
					}
				}
			case "name":
				if p.matchesRegexps("name", regexps) {
					ch.SetName("")
				}
			}
		}
	}
}

func (p *Processor) setFields(ch *Channel, fields []rules.SetFieldSpec) {
	for _, field := range fields {
		var buf bytes.Buffer
		_ = field.Template.Execute(&buf, map[string]any{
			"Channel": map[string]any{
				"Name":  ch.Name(),
				"Attrs": ch.Attrs(),
				"Tags":  ch.Tags(),
			},
		})
		p.setValue(ch, field.Type, field.Name, buf.String())
	}
}

func (p *Processor) setValue(ch *Channel, fieldType, fieldName, value string) {
	switch fieldType {
	case "attr":
		ch.SetAttr(fieldName, value)
	case "tag":
		ch.SetTag(fieldName, value)
	case "name":
		ch.SetName(value)
	}
}

func (p *Processor) expandNamedConditions(
	channelRules []rules.ChannelRule, namedConditions []rules.NamedCondition) []rules.ChannelRule {
	if len(namedConditions) == 0 {
		return channelRules
	}

	conditionMap := make(map[string][]rules.Condition)
	for _, namedCondition := range namedConditions {
		conditionMap[namedCondition.Name] = namedCondition.When
	}

	expandedRules := make([]rules.ChannelRule, len(channelRules))
	for i, rule := range channelRules {
		expandedRules[i] = rule
		expandedRules[i].When = p.expandConditions(rule.When, conditionMap)
	}

	return expandedRules
}

func (p *Processor) expandConditions(
	conditions []rules.Condition, conditionMap map[string][]rules.Condition) []rules.Condition {
	var expanded []rules.Condition

	for _, condition := range conditions {
		if condition.Ref != "" {
			if namedConditions, exists := conditionMap[condition.Ref]; exists {
				expandedNamed := p.expandConditions(namedConditions, conditionMap)
				expanded = append(expanded, expandedNamed...)
			}
			continue
		}

		condition.And = p.expandConditions(condition.And, conditionMap)
		condition.Or = p.expandConditions(condition.Or, conditionMap)
		condition.Not = p.expandConditions(condition.Not, conditionMap)
		expanded = append(expanded, condition)
	}

	return expanded
}
