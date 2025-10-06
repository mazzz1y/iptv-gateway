package channel

import (
	"bytes"
	"majmun/internal/config/common"
	"majmun/internal/config/rules/channel"
	"majmun/internal/listing/m3u8/store"
)

type Processor struct {
	clientName string
	rules      []*channel.Rule
}

func NewRulesProcessor(clientName string, rules []*channel.Rule) *Processor {
	return &Processor{
		clientName: clientName,
		rules:      rules,
	}
}

func (p *Processor) Apply(store *store.Store) {
	for _, ch := range store.All() {
		for _, rule := range p.rules {
			p.processChannelRule(ch, rule)
		}
	}
}

func (p *Processor) processChannelRule(ch *store.Channel, rule *channel.Rule) (stop bool) {
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

func (p *Processor) processSetField(ch *store.Channel, rule *channel.SetFieldRule) {
	if rule.Condition != nil && !p.matchesCondition(ch, *rule.Condition) {
		return
	}

	pl := ch.Playlist()
	tmplMap := map[string]any{
		"Channel": map[string]any{
			"Name":  ch.Name(),
			"Attrs": ch.Attrs(),
			"Tags":  ch.Tags(),
		},
		"Playlist": map[string]any{
			"Name":      pl.Name(),
			"IsProxied": pl.IsProxied(),
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

func (p *Processor) processRemoveField(ch *store.Channel, rule *channel.RemoveFieldRule) {
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

func (p *Processor) processRemoveChannel(ch *store.Channel, rule *channel.RemoveChannelRule) bool {
	if rule.Condition != nil && !p.matchesCondition(ch, *rule.Condition) {
		return false
	}
	ch.MarkRemoved()
	return true
}

func (p *Processor) processMarkHidden(ch *store.Channel, rule *channel.MarkHiddenRule) {
	if rule.Condition != nil && !p.matchesCondition(ch, *rule.Condition) {
		return
	}
	ch.MarkHidden()
}

func (p *Processor) matchesCondition(ch *store.Channel, condition common.Condition) bool {
	if condition.IsEmpty() {
		return true
	}

	fieldResult := p.evaluateField(ch, condition)

	var result bool
	if len(condition.And) > 0 {
		result = fieldResult && p.evaluateAnd(ch, condition.And)
	} else if len(condition.Or) > 0 {
		result = fieldResult && p.evaluateOr(ch, condition.Or)
	} else {
		result = fieldResult
	}

	if condition.Invert {
		result = !result
	}

	return result
}

func (p *Processor) evaluateField(ch *store.Channel, condition common.Condition) bool {
	hasFieldConditions := condition.Selector != nil || len(condition.Patterns) > 0 ||
		len(condition.Clients) > 0 || len(condition.Playlists) > 0

	if !hasFieldConditions {
		return true
	}

	if len(condition.Patterns) > 0 {
		fieldValue, ok := ch.GetFieldValue(condition.Selector)
		if !ok {
			return false
		}
		if !p.matchesRegexps(fieldValue, condition.Patterns) {
			return false
		}
	}

	if len(condition.Clients) > 0 && !p.matchesExactStrings(p.clientName, condition.Clients) {
		return false
	}

	if len(condition.Playlists) > 0 && !p.matchesExactStrings(ch.Playlist().Name(), condition.Playlists) {
		return false
	}

	return true
}

func (p *Processor) evaluateAnd(ch *store.Channel, conditions []common.Condition) bool {
	for _, sub := range conditions {
		if !p.matchesCondition(ch, sub) {
			return false
		}
	}
	return true
}

func (p *Processor) evaluateOr(ch *store.Channel, conditions []common.Condition) bool {
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
