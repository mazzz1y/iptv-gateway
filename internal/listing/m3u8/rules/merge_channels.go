package rules

import (
	configrules "iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/parser/m3u8"
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

		if bestTvgId, exists := best.GetAttr(m3u8.AttrTvgID); exists {
			for _, ch := range group {
				ch.SetAttr(m3u8.AttrTvgID, bestTvgId)
			}
		}

		for i, ch := range group {
			if ch == best { // Move the best channel to the front of the group
				group[0], group[i] = group[i], group[0]
				break
			}
		}

		if p.rule.SetField != nil {
			var finalValue string

			switch {
			case p.rule.SetField.NameTemplate != nil:
				finalValue = processSetField(best, p.rule.SetField.NameTemplate, baseName)
				for _, ch := range group {
					ch.SetName(finalValue)
				}

			case p.rule.SetField.AttrTemplate != nil:
				finalValue = processSetField(best, p.rule.SetField.AttrTemplate.Template, baseName)
				attrName := p.rule.SetField.AttrTemplate.Name
				for _, ch := range group {
					ch.SetAttr(attrName, finalValue)
				}

			case p.rule.SetField.TagTemplate != nil:
				finalValue = processSetField(best, p.rule.SetField.TagTemplate.Template, baseName)
				tagName := p.rule.SetField.TagTemplate.Name
				for _, ch := range group {
					ch.SetTag(tagName, finalValue)
				}
			}
		}
	}
}
