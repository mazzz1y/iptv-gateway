package rules

import (
	configrules "iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/rules/playlist"
	"regexp"
	"strings"
)

type RemoveDuplicatesProcessor struct {
	FieldType string
	FieldName string
	Patterns  []*regexp.Regexp
}

func NewRemoveDuplicatesActionProcessor(rule *playlist.RemoveDuplicatesRule) *RemoveDuplicatesProcessor {
	if len(rule.NamePatterns) > 0 {
		return &RemoveDuplicatesProcessor{
			FieldType: configrules.FieldTypeName,
			Patterns:  rule.NamePatterns.ToArray(),
		}
	}
	if rule.AttrPatterns != nil {
		return &RemoveDuplicatesProcessor{
			FieldType: configrules.FieldTypeAttr,
			FieldName: rule.AttrPatterns.Name,
			Patterns:  rule.AttrPatterns.Patterns.ToArray(),
		}
	}
	if rule.TagPatterns != nil {
		return &RemoveDuplicatesProcessor{
			FieldType: configrules.FieldTypeTag,
			FieldName: rule.TagPatterns.Name,
			Patterns:  rule.TagPatterns.Patterns.ToArray(),
		}
	}
	return &RemoveDuplicatesProcessor{}
}

func (p *RemoveDuplicatesProcessor) Apply(global, sub *Store) {
	grouped := make(map[string][]*Channel)
	for _, ch := range global.All() {
		key := p.extractBaseName(ch)
		grouped[key] = append(grouped[key], ch)
	}
	p.processDuplicateGroups(grouped, sub)
}

func (p *RemoveDuplicatesProcessor) extractBaseName(ch *Channel) string {
	name := p.getFieldValue(ch)
	if name == "" {
		return ""
	}
	for _, regex := range p.Patterns {
		name = regex.ReplaceAllString(name, "")
	}

	return strings.Join(strings.Fields(name), " ")
}

func (p *RemoveDuplicatesProcessor) getFieldValue(ch *Channel) string {
	switch p.FieldType {
	case configrules.FieldTypeAttr:
		if val, ok := ch.GetAttr(p.FieldName); ok {
			return val
		}
	case configrules.FieldTypeTag:
		if val, ok := ch.GetTag(p.FieldName); ok {
			return val
		}
	}
	return ch.Name()
}

func (p *RemoveDuplicatesProcessor) selectBestChannel(channels []*Channel) *Channel {
	if len(p.Patterns) == 0 {
		return channels[0]
	}
	for _, pattern := range p.Patterns {
		for _, ch := range channels {
			fv := p.getFieldValue(ch)
			if pattern.MatchString(fv) {
				return ch
			}
		}
	}
	return channels[0]
}

func (p *RemoveDuplicatesProcessor) processDuplicateGroups(groups map[string][]*Channel, sub *Store) {
	subMap := make(map[*Channel]bool)
	for _, ch := range sub.All() {
		subMap[ch] = true
	}
	for _, group := range groups {
		if len(group) <= 1 {
			continue
		}
		best := p.selectBestChannel(group)
		var subGroup []*Channel
		for _, ch := range group {
			if subMap[ch] {
				subGroup = append(subGroup, ch)
			}
		}
		if len(subGroup) == 0 {
			continue
		}
		bestInSub := false
		for _, ch := range subGroup {
			if ch == best {
				bestInSub = true
				break
			}
		}
		for _, ch := range subGroup {
			if !bestInSub || ch != best {
				ch.MarkRemoved()
			}
		}
	}
}
