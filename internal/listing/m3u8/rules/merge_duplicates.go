package rules

import (
	"bytes"
	configrules "iptv-gateway/internal/config/rules/playlist"

	"iptv-gateway/internal/config/common"
	"iptv-gateway/internal/parser/m3u8"
)

type MergeDuplicatesProcessor struct {
	rule *configrules.MergeDuplicatesRule
}

func NewMergeDuplicatesActionProcessor(rule *configrules.MergeDuplicatesRule) *MergeDuplicatesProcessor {
	return &MergeDuplicatesProcessor{rule: rule}
}

func (p *MergeDuplicatesProcessor) Apply(store *Store) {
	grouped := make(map[string][]*Channel)
	for _, ch := range store.All() {
		key, ok := extractBaseNameFromChannel(ch, p.rule.Selector, p.rule.Patterns)
		if !ok {
			continue
		}
		grouped[key] = append(grouped[key], ch)
	}
	p.processMergeGroups(grouped)
}

func (p *MergeDuplicatesProcessor) processMergeGroups(groups map[string][]*Channel) {
	for baseName, group := range groups {
		if len(group) <= 1 {
			continue
		}

		if !hasPatternVariationsGroup(group, p.rule.Selector, p.rule.Patterns) {
			continue
		}

		best := selectBestChannel(group, p.rule.Selector, p.rule.Patterns)

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

		if p.rule.FinalValue != nil {
			tmplMap := map[string]any{
				"Channel": map[string]any{
					"Name":  best.Name(),
					"Attrs": best.Attrs(),
					"Tags":  best.Tags(),
				},
				"BaseName": baseName,
			}

			var buf bytes.Buffer
			if err := p.rule.FinalValue.Template.ToTemplate().Execute(&buf, tmplMap); err == nil {
				finalValue := buf.String()

				for _, ch := range group {
					switch p.rule.FinalValue.Selector.Type {
					case common.SelectorName:
						ch.SetName(finalValue)
					case common.SelectorAttr:
						ch.SetAttr(p.rule.FinalValue.Selector.Value, finalValue)
					case common.SelectorTag:
						ch.SetTag(p.rule.FinalValue.Selector.Value, finalValue)
					}
				}
			}
		}
	}
}
