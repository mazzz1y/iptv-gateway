package rules

import (
	configrules "iptv-gateway/internal/config/rules"
)

type MergeChannelsProcessor struct {
	rule *configrules.MergeChannelsRule
}

func NewMergeChannelsActionProcessor(rule *configrules.MergeChannelsRule) *MergeChannelsProcessor {
	return &MergeChannelsProcessor{rule: rule}
}

func (p *MergeChannelsProcessor) Apply(store *Store) {
	grouped := make(map[string][]*Channel)
	for _, ch := range store.All() {
		key := p.extractBaseName(ch)
		grouped[key] = append(grouped[key], ch)
	}
	p.processMergeGroups(grouped)
}

func (p *MergeChannelsProcessor) extractBaseName(ch *Channel) string {
	fv := getFieldValue(ch, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)
	patterns := getPatterns(p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)
	return extractBaseName(fv, patterns)
}

func (p *MergeChannelsProcessor) processMergeGroups(groups map[string][]*Channel) {
	for baseName, group := range groups {
		if len(group) <= 1 {
			continue
		}

		if !hasPatternVariationsGroup(group, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns) {
			continue
		}

		best := selectBestChannel(group, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)

		var finalValue string
		if p.rule.SetField != nil {
			finalValue = processSetField(best, p.rule.SetField, baseName)
		} else {
			finalValue = getFieldValue(best, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)
		}

		// Move best channel to front of group
		for i, ch := range group {
			if ch == best {
				group[0], group[i] = group[i], group[0]
				break
			}
		}

		for _, ch := range group {
			setFieldValue(ch, finalValue, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)
		}
	}
}
