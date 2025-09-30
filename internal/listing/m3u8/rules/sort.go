package rules

import (
	"iptv-gateway/internal/config/rules/playlist"
	"sort"
)

type SortProcessor struct {
	rule *playlist.Sort
}

func NewSortProcessor(rule *playlist.Sort) *SortProcessor {
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
			iVal, iOk := getSelectorFieldValue(channels[i], sp.rule.Selector)
			jVal, jOk := getSelectorFieldValue(channels[j], sp.rule.Selector)
			if !iOk {
				return false
			}
			if !jOk {
				return true
			}
			return iVal < jVal
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
			iVal, iOk := getSelectorFieldValue(groupChannels[i], sp.rule.Selector)
			jVal, jOk := getSelectorFieldValue(groupChannels[j], sp.rule.Selector)
			if !iOk {
				return false
			}
			if !jOk {
				return true
			}
			return iVal < jVal
		})
		sortedChannels = append(sortedChannels, groupChannels...)
	}

	store.Replace(sortedChannels)
}

func (sp *SortProcessor) getGroupKey(ch *Channel) string {
	if sp.rule.GroupBy == nil {
		return ""
	}
	if sp.rule.GroupBy.Selector != nil {
		val, _ := getSelectorFieldValue(ch, sp.rule.GroupBy.Selector)
		return val
	}
	return ""
}

func (sp *SortProcessor) getGroupPriority(groupValue string) int {
	if sp.rule.GroupBy == nil || sp.rule.GroupBy.Order == nil || len(*sp.rule.GroupBy.Order) == 0 {
		return 0
	}

	for i, pattern := range *sp.rule.GroupBy.Order {
		if pattern != nil && pattern.String() != "" && pattern.MatchString(groupValue) {
			return i
		}
	}

	for i, pattern := range *sp.rule.GroupBy.Order {
		if pattern != nil && pattern.String() == "" {
			return i
		}
	}

	return len(*sp.rule.GroupBy.Order)
}

func (sp *SortProcessor) getChannelPriority(ch *Channel) int {
	if sp.rule.Order == nil || len(*sp.rule.Order) == 0 {
		return 0
	}

	field, ok := getSelectorFieldValue(ch, sp.rule.Selector)
	if !ok {
		return len(*sp.rule.Order)
	}

	for i, pattern := range *sp.rule.Order {
		if pattern != nil && pattern.String() != "" && pattern.MatchString(field) {
			return i
		}
	}

	for i, pattern := range *sp.rule.Order {
		if pattern != nil && pattern.String() == "" {
			return i
		}
	}

	return len(*sp.rule.Order)
}
