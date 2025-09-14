package rules

import (
	"iptv-gateway/internal/config/rules/playlist"
	"regexp"
	"sort"
)

type SortProcessor struct {
	rule *playlist.SortRule
}

func NewSortProcessor(rule *playlist.SortRule) *SortProcessor {
	return &SortProcessor{rule: rule}
}

func (sp *SortProcessor) Apply(store *Store) {
	channels := store.All()
	if len(channels) <= 1 {
		return
	}

	if sp.rule.GroupBy == nil {
		sort.Slice(channels, func(i, j int) bool {
			iPriority := sp.getChannelPriority(channels[i])
			jPriority := sp.getChannelPriority(channels[j])
			if iPriority != jPriority {
				return iPriority < jPriority
			}
			return sp.getChannelSortValue(channels[i]) < sp.getChannelSortValue(channels[j])
		})
		store.Replace(channels)
		return
	}

	groups := make(map[string][]*Channel)
	for _, ch := range channels {
		groupKey := sp.getGroupKey(ch)
		groups[groupKey] = append(groups[groupKey], ch)
	}

	var sortedChannels []*Channel
	groupNames := make([]string, 0, len(groups))
	for name := range groups {
		groupNames = append(groupNames, name)
	}

	sort.Slice(groupNames, func(i, j int) bool {
		iPriority := sp.getGroupPriority(groupNames[i])
		jPriority := sp.getGroupPriority(groupNames[j])
		if iPriority != jPriority {
			return iPriority < jPriority
		}
		return groupNames[i] < groupNames[j]
	})

	for _, groupName := range groupNames {
		groupChannels := groups[groupName]
		sort.Slice(groupChannels, func(i, j int) bool {
			iPriority := sp.getChannelPriority(groupChannels[i])
			jPriority := sp.getChannelPriority(groupChannels[j])
			if iPriority != jPriority {
				return iPriority < jPriority
			}
			return sp.getChannelSortValue(groupChannels[i]) < sp.getChannelSortValue(groupChannels[j])
		})
		sortedChannels = append(sortedChannels, groupChannels...)
	}

	store.Replace(sortedChannels)
}

func (sp *SortProcessor) getGroupKey(ch *Channel) string {
	if sp.rule.GroupBy == nil {
		return ""
	}
	if sp.rule.GroupBy.Attr != "" {
		if value, ok := ch.GetAttr(sp.rule.GroupBy.Attr); ok {
			return value
		}
	}
	if sp.rule.GroupBy.Tag != "" {
		if value, ok := ch.GetTag(sp.rule.GroupBy.Tag); ok {
			return value
		}
	}
	return ""
}

func (sp *SortProcessor) getGroupPriority(groupValue string) int {
	if sp.rule.GroupBy == nil || sp.rule.GroupBy.Order == nil || len(*sp.rule.GroupBy.Order) == 0 {
		return 0
	}

	for i, pattern := range *sp.rule.GroupBy.Order {
		if pattern != "" && sp.matchesPattern(groupValue, pattern) {
			return i
		}
	}

	for i, pattern := range *sp.rule.GroupBy.Order {
		if pattern == "" {
			return i
		}
	}

	return len(*sp.rule.GroupBy.Order)
}

func (sp *SortProcessor) getChannelPriority(ch *Channel) int {
	if sp.rule.Order == nil || len(*sp.rule.Order) == 0 {
		return 0
	}

	channelValue := sp.getChannelSortValue(ch)

	for i, pattern := range *sp.rule.Order {
		if pattern != "" && sp.matchesPattern(channelValue, pattern) {
			return i
		}
	}

	for i, pattern := range *sp.rule.Order {
		if pattern == "" {
			return i
		}
	}

	return len(*sp.rule.Order)
}

func (sp *SortProcessor) getChannelSortValue(ch *Channel) string {
	if sp.rule.Attr != "" {
		if value, ok := ch.GetAttr(sp.rule.Attr); ok {
			return value
		}
	}

	if sp.rule.Tag != "" {
		if value, ok := ch.GetAttr(sp.rule.Tag); ok {
			return value
		}
	}

	return ch.Name()
}

func (sp *SortProcessor) matchesPattern(value, pattern string) bool {
	if pattern == "" {
		return true
	}

	if re, err := regexp.Compile(pattern); err == nil {
		return re.MatchString(value)
	}

	return value == pattern
}
