package rules

import (
	"bytes"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Processor struct {
	clientName   string
	channelRules []*rules.Rule
	storeRules   []*rules.Rule
}

func NewProcessor(clientName string, globalRules []*rules.Rule) *Processor {
	var channelRules []*rules.Rule
	var storeRules []*rules.Rule

	for _, rule := range globalRules {
		switch rule.Type {
		case rules.StoreRule:
			storeRules = append(storeRules, rule)
		case rules.ChannelRule:
			channelRules = append(channelRules, rule)
		}
	}

	return &Processor{
		clientName:   clientName,
		channelRules: channelRules,
		storeRules:   storeRules,
	}
}

func (p *Processor) Process(store *Store) {
	p.processTrackRules(store)
	p.processStoreRules(store)
}

func (p *Processor) processStoreRules(store *Store) {
	for _, rule := range p.storeRules {
		if rule.MergeChannels != nil && evaluateStoreWhenCondition(rule.MergeChannels.When, p.clientName) {
			processor := NewMergeChannelsActionProcessor(rule.MergeChannels)
			processor.Apply(store)
		}
		if rule.RemoveDuplicates != nil && evaluateStoreWhenCondition(rule.RemoveDuplicates.When, p.clientName) {
			processor := NewRemoveDuplicatesActionProcessor(rule.RemoveDuplicates)
			processor.Apply(store)
		}
		if rule.SortRule != nil {
			processor := NewSortProcessor(rule.SortRule)
			processor.Apply(store)
		}
	}
}

func (p *Processor) processTrackRules(store *Store) {
	for _, ch := range store.All() {
		for _, rule := range p.channelRules {
			p.processChannelRule(ch, rule)
		}
	}
}

func (p *Processor) processChannelRule(ch *Channel, rule *rules.Rule) (stop bool) {
	if rule.SetField != nil {
		p.processSetField(ch, rule.SetField)
		stop = false
	} else if rule.RemoveField != nil {
		p.processRemoveField(ch, rule.RemoveField)
		stop = false
	} else if rule.RemoveChannel != nil {
		stop = p.processRemoveChannel(ch, rule.RemoveChannel)
	} else if rule.MarkHidden != nil {
		p.processMarkHidden(ch, rule.MarkHidden)
		stop = false
	}
	return
}

func (p *Processor) processSetField(ch *Channel, rule *rules.SetFieldRule) {
	if rule.When != nil && !p.matchesCondition(ch, *rule.When) {
		return
	}

	tmplMap := map[string]any{
		"Channel": map[string]any{
			"Name":  ch.Name(),
			"Attrs": ch.Attrs(),
			"Tags":  ch.Tags(),
		},
	}
	var buf bytes.Buffer

	switch {
	case rule.NameTemplate != nil:
		if err := rule.NameTemplate.ToTemplate().Execute(&buf, tmplMap); err != nil {
			return
		}
		ch.SetName(buf.String())
	case rule.AttrTemplate != nil:
		if err := rule.AttrTemplate.Template.ToTemplate().Execute(&buf, tmplMap); err != nil {
			return
		}
		ch.SetAttr(rule.AttrTemplate.Name, buf.String())
	case rule.TagTemplate != nil:
		if err := rule.TagTemplate.Template.ToTemplate().Execute(&buf, tmplMap); err != nil {
			return
		}
		ch.SetTag(rule.TagTemplate.Name, buf.String())
	}
}

func (p *Processor) processRemoveField(ch *Channel, rule *rules.RemoveFieldRule) {
	if rule.When != nil && !p.matchesCondition(ch, *rule.When) {
		return
	}

	switch {
	case rule.AttrPatterns != nil:
		for attrKey := range ch.Attrs() {
			for _, pattern := range rule.AttrPatterns {
				if pattern.MatchString(attrKey) {
					ch.DeleteAttr(attrKey)
					break
				}
			}
		}
	case rule.TagPatterns != nil:
		for tagKey := range ch.Tags() {
			for _, pattern := range rule.TagPatterns {
				if pattern.MatchString(tagKey) {
					ch.DeleteTag(tagKey)
					break
				}
			}
		}
	}
}

func (p *Processor) processRemoveChannel(ch *Channel, rule *rules.RemoveChannelRule) bool {
	if rule.When != nil && !p.matchesCondition(ch, *rule.When) {
		return false
	}
	ch.MarkRemoved()
	return true
}

func (p *Processor) processMarkHidden(ch *Channel, rule *rules.MarkHiddenRule) {
	if rule.When != nil && !p.matchesCondition(ch, *rule.When) {
		return
	}
	ch.MarkHidden()
}

func (p *Processor) matchesCondition(ch *Channel, condition types.Condition) bool {
	if condition.IsEmpty() {
		return true
	}

	fieldResult := p.evaluateFieldCondition(ch, condition)

	var result bool
	if len(condition.And) > 0 {
		result = fieldResult && p.evaluateAndConditions(ch, condition.And)
	} else if len(condition.Or) > 0 {
		result = fieldResult && p.evaluateOrConditions(ch, condition.Or)
	} else {
		result = fieldResult
	}

	if condition.Invert {
		result = !result
	}

	return result
}

func (p *Processor) evaluateFieldCondition(ch *Channel, condition types.Condition) bool {
	hasFieldConditions := condition.NamePatterns != nil || condition.Attr != nil || condition.Tag != nil ||
		len(condition.Clients) > 0 || len(condition.Playlists) > 0

	if !hasFieldConditions {
		return true
	}

	if condition.NamePatterns != nil && !p.matchesRegexps(ch.Name(), condition.NamePatterns) {
		return false
	}

	if condition.Attr != nil {
		actual, exists := ch.GetAttr(condition.Attr.Name)
		if !exists || !p.matchesRegexps(actual, condition.Attr.Patterns) {
			return false
		}
	}

	if condition.Tag != nil {
		actual, exists := ch.GetTag(condition.Tag.Name)
		if !exists || !p.matchesRegexps(actual, condition.Tag.Patterns) {
			return false
		}
	}

	if len(condition.Clients) > 0 && !p.matchesExactStrings(p.clientName, condition.Clients) {
		return false
	}

	if len(condition.Playlists) > 0 && !p.matchesExactStrings(ch.Subscription().Name(), condition.Playlists) {
		return false
	}

	return true
}

func (p *Processor) evaluateAndConditions(ch *Channel, conditions types.ConditionList) bool {
	for _, sub := range conditions {
		if !p.matchesCondition(ch, sub) {
			return false
		}
	}
	return true
}

func (p *Processor) evaluateOrConditions(ch *Channel, conditions types.ConditionList) bool {
	for _, sub := range conditions {
		if p.matchesCondition(ch, sub) {
			return true
		}
	}
	return false
}

func (p *Processor) matchesRegexps(value string, regexps types.RegexpArr) bool {
	for _, re := range regexps {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func (p *Processor) matchesExactStrings(value string, strings types.StringOrArr) bool {
	for _, str := range strings {
		if value == str {
			return true
		}
	}
	return false
}

func (p *Processor) containsStoreRule(rule *rules.Rule) bool {
	for _, existing := range p.storeRules {
		if existing == rule {
			return true
		}
	}
	return false
}
