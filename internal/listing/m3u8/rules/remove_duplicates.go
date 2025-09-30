package rules

import (
	"bytes"
	"iptv-gateway/internal/config/common"
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
		key, ok := extractBaseNameFromChannel(ch, p.rule.Selector, p.rule.Patterns)
		if !ok {
			continue
		}
		grouped[key] = append(grouped[key], ch)
	}
	p.processDuplicateGroups(grouped)
}

func (p *RemoveDuplicatesProcessor) processDuplicateGroups(groups map[string][]*Channel) {
	for baseName, group := range groups {
		if len(group) <= 1 {
			continue
		}

		if !hasPatternVariationsGroup(group, p.rule.Selector, p.rule.Patterns) {
			continue
		}

		best := selectBestChannel(group, p.rule.Selector, p.rule.Patterns)
		for _, ch := range group {
			if ch == best {
				if p.rule.FinalValue != nil {
					tmplMap := map[string]any{
						"Channel": map[string]any{
							"Name":  ch.Name(),
							"Attrs": ch.Attrs(),
							"Tags":  ch.Tags(),
						},
						"BaseName": baseName,
					}

					var buf bytes.Buffer
					if err := p.rule.FinalValue.Template.ToTemplate().Execute(&buf, tmplMap); err == nil {
						finalValue := buf.String()

						if p.rule.FinalValue.Selector != nil {
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
			} else {
				ch.MarkRemoved()
			}
		}
	}
}
