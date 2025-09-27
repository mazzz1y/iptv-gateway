package rules

import (
	"iptv-gateway/internal/config/rules"
)

type RemoveDuplicatesProcessor struct {
	rule *rules.RemoveDuplicatesRule
}

func NewRemoveDuplicatesActionProcessor(rule *rules.RemoveDuplicatesRule) *RemoveDuplicatesProcessor {
	return &RemoveDuplicatesProcessor{rule: rule}
}

func (p *RemoveDuplicatesProcessor) Apply(global *Store) {
	grouped := make(map[string][]*Channel)
	for _, ch := range global.All() {
		key := p.extractBaseName(ch)
		grouped[key] = append(grouped[key], ch)
	}
	p.processDuplicateGroups(grouped)
}

func (p *RemoveDuplicatesProcessor) extractBaseName(ch *Channel) string {
	fv := getFieldValue(ch, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)
	patterns := getPatterns(p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)
	return extractBaseName(fv, patterns)
}

func (p *RemoveDuplicatesProcessor) processDuplicateGroups(groups map[string][]*Channel) {
	for baseName, group := range groups {
		if len(group) <= 1 {
			continue
		}

		if !hasPatternVariationsGroup(group, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns) {
			continue
		}

		best := selectBestChannel(group, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)
		for _, ch := range group {
			if ch == best {
				if p.rule.SetField != nil {
					finalValue := processSetField(ch, p.rule.SetField, baseName)
					setFieldValue(ch, finalValue, p.rule.NamePatterns, p.rule.AttrPatterns, p.rule.TagPatterns)
				}
			} else {
				ch.MarkRemoved()
			}
		}
	}
}
