package rules

import (
	"bytes"
	"iptv-gateway/internal/config/common"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/listing"
)

type Processor struct {
	clientName   string
	channelRules []*rules.ChannelRule
	storeRules   []*rules.PlaylistRule
}

func NewProcessor(clientName string, channelRules []*rules.ChannelRule, playlistRules []*rules.PlaylistRule) *Processor {
	return &Processor{
		clientName:   clientName,
		channelRules: channelRules,
		storeRules:   playlistRules,
	}
}

func (p *Processor) Process(store *Store) {
	p.processTrackRules(store)
	p.processStoreRules(store)
}

func (p *Processor) processStoreRules(store *Store) {
	for _, rule := range p.storeRules {
		if rule.MergeChannels != nil && evaluateStoreCondition(rule.MergeChannels.Condition, p.clientName) {
			processor := NewMergeDuplicatesActionProcessor(rule.MergeChannels)
			processor.Apply(store)
		}
		if rule.RemoveDuplicates != nil && evaluateStoreCondition(rule.RemoveDuplicates.Condition, p.clientName) {
			processor := NewRemoveDuplicatesActionProcessor(rule.RemoveDuplicates)
			processor.Apply(store)
		}
		if rule.SortRule != nil && evaluateStoreCondition(rule.SortRule.Condition, p.clientName) {
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
		p.ensureTvgID(ch)
	}
}

func (p *Processor) ensureTvgID(ch *Channel) {
	if tvgID, exists := ch.GetAttr("tvg-id"); exists && tvgID != "" {
		return
	}
	if tvgName, exists := ch.GetAttr("tvg-name"); exists && tvgName != "" {
		ch.SetAttr("tvg-id", listing.GenerateHashID(tvgName))
	} else {
		ch.SetAttr("tvg-id", listing.GenerateHashID(ch.Name()))
	}
}

func (p *Processor) processChannelRule(ch *Channel, rule *rules.ChannelRule) (stop bool) {
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
	if rule.Condition != nil && !p.matchesCondition(ch, *rule.Condition) {
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

	if err := rule.Template.ToTemplate().Execute(&buf, tmplMap); err != nil {
		return
	}

	value := buf.String()
	switch rule.Selector.Type {
	case common.SelectorName:
		ch.SetName(value)
	case common.SelectorAttr:
		ch.SetAttr(rule.Selector.Value, value)
	case common.SelectorTag:
		ch.SetTag(rule.Selector.Value, value)
	}
}

func (p *Processor) processRemoveField(ch *Channel, rule *rules.RemoveFieldRule) {
	if rule.Condition != nil && !p.matchesCondition(ch, *rule.Condition) {
		return
	}

	switch rule.Selector.Type {
	case common.SelectorAttr:
		for attrKey := range ch.Attrs() {
			if attrKey == rule.Selector.Value {
				ch.DeleteAttr(attrKey)
				break
			}
		}
	case common.SelectorTag:
		for tagKey := range ch.Tags() {
			if tagKey == rule.Selector.Value {
				ch.DeleteAttr(tagKey)
				break
			}
		}
	}
}

func (p *Processor) processRemoveChannel(ch *Channel, rule *rules.RemoveChannelRule) bool {
	if rule.Condition != nil && !p.matchesCondition(ch, *rule.Condition) {
		return false
	}
	ch.MarkRemoved()
	return true
}

func (p *Processor) processMarkHidden(ch *Channel, rule *rules.MarkHiddenRule) {
	if rule.Condition != nil && !p.matchesCondition(ch, *rule.Condition) {
		return
	}
	ch.MarkHidden()
}

func (p *Processor) matchesCondition(ch *Channel, condition common.Condition) bool {
	if condition.IsEmpty() {
		return true
	}

	fieldResult := p.evaluateConditionFieldCondition(ch, condition)

	var result bool
	if len(condition.And) > 0 {
		result = fieldResult && p.evaluateConditionAndConditions(ch, condition.And)
	} else if len(condition.Or) > 0 {
		result = fieldResult && p.evaluateConditionOrConditions(ch, condition.Or)
	} else {
		result = fieldResult
	}

	if condition.Invert {
		result = !result
	}

	return result
}

func (p *Processor) evaluateConditionFieldCondition(ch *Channel, condition common.Condition) bool {
	hasFieldConditions := condition.Selector != nil || len(condition.Patterns) > 0 || len(condition.Clients) > 0 || len(condition.Playlists) > 0

	if !hasFieldConditions {
		return true
	}

	if len(condition.Patterns) > 0 {
		fieldValue := getSelectorFieldValue(ch, condition.Selector)
		if !p.matchesRegexps(fieldValue, condition.Patterns) {
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

func (p *Processor) evaluateConditionAndConditions(ch *Channel, conditions []common.Condition) bool {
	for _, sub := range conditions {
		if !p.matchesCondition(ch, sub) {
			return false
		}
	}
	return true
}

func (p *Processor) evaluateConditionOrConditions(ch *Channel, conditions []common.Condition) bool {
	for _, sub := range conditions {
		if p.matchesCondition(ch, sub) {
			return true
		}
	}
	return false
}

func (p *Processor) matchesRegexps(value string, regexps common.RegexpArr) bool {
	for _, re := range regexps {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func (p *Processor) matchesExactStrings(value string, strings common.StringOrArr) bool {
	for _, str := range strings {
		if value == str {
			return true
		}
	}
	return false
}

func (p *Processor) containsStoreRule(rule *rules.PlaylistRule) bool {
	for _, existing := range p.storeRules {
		if existing == rule {
			return true
		}
	}
	return false
}
