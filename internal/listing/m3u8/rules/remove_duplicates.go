package rules

import (
	"iptv-gateway/internal/config/rules/playlist"
	"regexp"
	"strings"
)

type RemoveDuplicatesProcessor struct {
	rule *playlist.RemoveDuplicatesRule
}

func NewRemoveDuplicatesActionProcessor(rule *playlist.RemoveDuplicatesRule) *RemoveDuplicatesProcessor {
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
	name := p.getFieldValue(ch)
	if name == "" {
		return ""
	}
	patterns := p.getPatterns()
	for _, regex := range patterns {
		name = regex.ReplaceAllString(name, "")
	}

	return strings.Join(strings.Fields(name), " ")
}

func (p *RemoveDuplicatesProcessor) getFieldValue(ch *Channel) string {
	if len(p.rule.NamePatterns) > 0 {
		return ch.Name()
	}
	if p.rule.AttrPatterns != nil {
		if val, ok := ch.GetAttr(p.rule.AttrPatterns.Name); ok {
			return val
		}
	}
	if p.rule.TagPatterns != nil {
		if val, ok := ch.GetTag(p.rule.TagPatterns.Name); ok {
			return val
		}
	}
	return ch.Name()
}

func (p *RemoveDuplicatesProcessor) getPatterns() []*regexp.Regexp {
	if len(p.rule.NamePatterns) > 0 {
		return p.rule.NamePatterns.ToArray()
	}
	if p.rule.AttrPatterns != nil {
		return p.rule.AttrPatterns.Patterns.ToArray()
	}
	if p.rule.TagPatterns != nil {
		return p.rule.TagPatterns.Patterns.ToArray()
	}
	return nil
}

func (p *RemoveDuplicatesProcessor) selectBestChannel(channels []*Channel) *Channel {
	patterns := p.getPatterns()
	if len(patterns) == 0 {
		return channels[0]
	}
	for _, pattern := range patterns {
		for _, ch := range channels {
			fv := p.getFieldValue(ch)
			if pattern.MatchString(fv) {
				return ch
			}
		}
	}
	return channels[0]
}

func (p *RemoveDuplicatesProcessor) processDuplicateGroups(groups map[string][]*Channel) {
	for baseName, group := range groups {
		if len(group) <= 1 {
			continue
		}
		best := p.selectBestChannel(group)
		for _, ch := range group {
			if ch == best {
				if p.rule.TrimPattern {
					ch.SetName(baseName)
				}
			} else {
				ch.MarkRemoved()
			}
		}
	}
}
