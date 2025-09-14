package rules

import (
	"bytes"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/rules/channel"
	"iptv-gateway/internal/config/rules/playlist"
	"iptv-gateway/internal/config/types"
	"iptv-gateway/internal/listing"
)

type Processor struct {
	subscriptionChannelRulesMap  map[listing.Playlist][]channel.Rule
	subscriptionPlaylistRulesMap map[listing.Playlist][]playlist.Rule
}

func NewProcessor() *Processor {
	return &Processor{
		subscriptionChannelRulesMap:  make(map[listing.Playlist][]channel.Rule),
		subscriptionPlaylistRulesMap: make(map[listing.Playlist][]playlist.Rule),
	}
}

func (p *Processor) AddSubscription(sub listing.Playlist) {
	p.subscriptionChannelRulesMap[sub] = append(p.subscriptionChannelRulesMap[sub], sub.ChannelRules()...)
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
			for _, rule := range subRules {
				p.processChannelRule(ch, rule)
			}
		}
	}
}

func (p *Processor) processChannelWithRules(ch *Channel, rules []channel.Rule) {
	for _, rule := range rules {
		if p.processChannelRule(ch, rule) {
			return
		}
	}
}

func (p *Processor) processChannelRule(ch *Channel, rule channel.Rule) (stop bool) {
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

func (p *Processor) processSetField(ch *Channel, rule *channel.SetFieldRule) {
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

func (p *Processor) processRemoveField(ch *Channel, rule *channel.RemoveFieldRule) {
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

func (p *Processor) processRemoveChannel(ch *Channel, rule *channel.RemoveChannelRule) bool {
	if rule.When != nil && !p.matchesCondition(ch, *rule.When) {
		return false
	}
	ch.MarkRemoved()
	return true
}

func (p *Processor) processMarkHidden(ch *Channel, rule *channel.MarkHiddenRule) {
	if rule.When != nil && !p.matchesCondition(ch, *rule.When) {
		return
	}
	ch.MarkHidden()
}

func (p *Processor) processRemoveDuplicatesRules(global *Store) {
	for sub, rul := range p.subscriptionPlaylistRulesMap {
		subStore := NewStore()
		for _, ch := range global.All() {
			if ch.Subscription() == sub {
				subStore.Add(ch)
			}
		}

		for _, rule := range rul {
			if rule.RemoveDuplicates != nil {
				processor := NewRemoveDuplicatesActionProcessor(rule.RemoveDuplicates)
				processor.Apply(global, subStore)
			}
		}
	}
}

func (p *Processor) matchesCondition(ch *Channel, condition rules.Condition) bool {
	if condition.IsEmpty() {
		return true
	}

	result := p.evaluateFieldCondition(ch, condition)

	if len(condition.And) > 0 {
		result = p.evaluateAndConditions(ch, condition.And)
	} else if len(condition.Or) > 0 {
		result = p.evaluateOrConditions(ch, condition.Or)
	}

	if condition.Invert {
		result = !result
	}

	return result
}

func (p *Processor) evaluateFieldCondition(ch *Channel, condition rules.Condition) bool {
	if condition.NamePatterns != nil {
		return p.matchesRegexps(ch.Name(), condition.NamePatterns)
	}
	if condition.Attr != nil {
		if actual, exists := ch.GetAttr(condition.Attr.Name); exists {
			return p.matchesRegexps(actual, condition.Attr.Patterns)
		}
	}
	if condition.Tag != nil {
		if actual, exists := ch.GetTag(condition.Tag.Name); exists {
			return p.matchesRegexps(actual, condition.Tag.Patterns)
		}
	}
	return false
}

func (p *Processor) evaluateAndConditions(ch *Channel, conditions rules.ConditionList) bool {
	for _, sub := range conditions {
		if !p.matchesCondition(ch, sub) {
			return false
		}
	}
	return true
}

func (p *Processor) evaluateOrConditions(ch *Channel, conditions rules.ConditionList) bool {
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
